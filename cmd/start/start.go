package start

import (
	"errors"
	"fmt"
	"minter-sentinel/config"
	"minter-sentinel/services/minter/node"
	"minter-sentinel/services/prometheus"
	"minter-sentinel/services/telegram"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var NoValidatorsSignedYet = errors.New("no validators signed")

type Command struct {
	log    *logrus.Logger
	config *config.Config

	wg sync.WaitGroup

	dryRun         bool
	missedBlocks   []int
	lastBlock      int
	controlAddress string

	minter     *node.Service
	telegram   *tgbotapi.BotAPI
	prometheus *prometheus.Service
}

func New(log *logrus.Logger, config *config.Config) *Command {
	return &Command{
		log:    log,
		config: config,
	}
}

func (cmd *Command) Command() *cli.Command {
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:     "dry-run",
			Required: false,
			Value:    false,
			Usage:    "Don't send transaction to set masternode off",
		},
	}

	return &cli.Command{
		Name:  "start",
		Usage: "Start watching for missed blocks",
		Flags: flags,
		Action: func(ctx *cli.Context) error {
			cmd.dryRun = ctx.Bool("dry-run")

			if !cmd.dryRun && len(cmd.config.Minter.TransactionOff) == 0 && len(cmd.config.Minter.Seeds) == 0 {
				return errors.New("`transaction_off` or `seeds` are not set in configuration file")
			}

			if len(cmd.config.Minter.NodeApi) == 0 {
				return errors.New("define at least one node_api in configuration file")
			}

			if n, err := node.New(cmd.config.Minter.NodeApi, cmd.config.Minter.Testnet, cmd.log); err != nil {
				return err
			} else {
				if err := n.Ping(); err != nil {
					return err
				}

				cmd.minter = n
			}

			if cmd.config.Prometheus.Enabled {
				if p, err := prometheus.New(cmd.config.Prometheus.Address, cmd.log); err != nil {
					return err
				} else {
					cmd.prometheus = p
					cmd.prometheus.SetMissedBlocksThreshold(cmd.config.Minter.MissedBlocksThreshold)
					cmd.prometheus.SetSleep(cmd.config.Minter.Sleep)

					go func() {
						if err := cmd.prometheus.Start(); err != nil {
							cmd.newLogEntry(cmd.lastBlock).Fatalln(err)
						}
					}()
				}
			}

			if len(cmd.config.Telegram.Token) > 0 {
				b, err := telegram.New(cmd.config.Telegram.Token)

				if err != nil {
					return err
				}

				cmd.telegram = b
			} else {
				cmd.log.Warn("Telegram token not set. Notifications will not be sent")
			}

			candidate, err := cmd.minter.GetCandidate(cmd.config.Minter.PublicKey)

			if err != nil {
				return err
			}

			lastBlock, err := cmd.lastBlockHeight()

			if err != nil {
				return err
			}

			if candidate.JailedUntil > 0 && candidate.JailedUntil > lastBlock {
				return errors.New(
					fmt.Sprintf(
						"candidate is jailed until block %d, current block: %d, blocks until unjail: %d (approx. %d sec.)",
						candidate.JailedUntil,
						lastBlock,
						candidate.JailedUntil-lastBlock,
						(candidate.JailedUntil-lastBlock)*5,
					),
				)
			}

			if !candidate.Validator {
				return errors.New("candidate is not a validator yet")
			}

			if candidate.Status != 2 {
				return errors.New("candidate is not online")
			}

			signed, err := cmd.isSigned(lastBlock)

			if err != nil {
				return err
			}

			if !signed {
				return errors.New("last block is not signed by the validator, start watcher when validator starts signing blocks")
			}

			cmd.lastBlock = lastBlock
			cmd.controlAddress = candidate.ControlAddress

			return cmd.run()
		},
	}
}

func (cmd *Command) run() error {
	cmd.newLogEntry(cmd.lastBlock).
		WithField("missed_blocks_threshold", cmd.config.Minter.MissedBlocksThreshold).
		WithField("sleep", cmd.config.Minter.Sleep).
		WithField("missed_block_remove_after", cmd.config.Minter.MissedBlockRemoveAfter).
		WithField("control_address", cmd.controlAddress).
		Println("Watcher started")

	ticker := time.NewTicker(time.Duration(cmd.config.Minter.Sleep) * time.Second)

	turnOff := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				cmd.cleanupMissedBlocks()

				nextHeight := cmd.lastBlock + 1

				signed, err := cmd.isSigned(nextHeight)

				if err != nil {
					if _, ok := err.(*node.BlockNotFound); ok {
						cmd.newLogEntry(nextHeight).Debugln("Block not created yet.")
						continue
					}

					if errors.Is(err, NoValidatorsSignedYet) {
						continue
					}

					go cmd.sendBotMessage(fmt.Sprintf("?????? Failed to detect if block is signed: %s", err))

					turnOff <- true

					return
				}

				cmd.lastBlock = nextHeight

				if signed {
					go func() {
						if cmd.prometheus != nil {
							cmd.prometheus.BlocksSignedIncrement()
						}
					}()

					cmd.newLogEntry(nextHeight).Println("Block signed")

					continue
				}

				go func() {
					if cmd.prometheus != nil {
						go cmd.prometheus.BlocksMissedIncrement()
					}
				}()

				cmd.missedBlocks = append(cmd.missedBlocks, nextHeight)

				if len(cmd.missedBlocks) > 0 && len(cmd.missedBlocks) < cmd.config.Minter.MissedBlocksThreshold {
					cmd.newLogEntry(nextHeight).Warnln("Block missed")

					go cmd.sendBotMessage(fmt.Sprintf("?????? Block %d missed [%d/%d]", nextHeight, len(cmd.missedBlocks), cmd.config.Minter.MissedBlocksThreshold))
				}

				if len(cmd.missedBlocks) >= cmd.config.Minter.MissedBlocksThreshold {
					cmd.newLogEntry(nextHeight).Errorln("Missed blocks threshold exceeded")

					go cmd.sendBotMessage(fmt.Sprintf("???? Block %d missed [%d/%d]", nextHeight, len(cmd.missedBlocks), cmd.config.Minter.MissedBlocksThreshold))

					turnOff <- true

					return
				}
			case <-turnOff:
				ticker.Stop()
				return
			}
		}
	}()

	if v := <-turnOff; v {
		go cmd.sendBotMessage("???? Sending transaction to turn off masternode")

		if err := cmd.turnOffMasternode(); err != nil {
			cmd.newLogEntry(cmd.lastBlock).Errorln("Failed to turn off masternode", err)

			go cmd.sendBotMessage("???? Failed to turn off masternode")

			return err
		}

		go cmd.sendBotMessage("???? Masternode is off")
	}

	cmd.wg.Wait()

	return nil
}

func (cmd *Command) sendBotMessage(message string) {
	if cmd.telegram == nil {
		return
	}

	cmd.wg.Add(1)

	for _, id := range cmd.config.Telegram.Admins {
		msg := tgbotapi.NewMessage(int64(id), message)

		_, _ = cmd.telegram.Send(msg)
	}

	cmd.wg.Done()
}

func (cmd *Command) turnOffMasternode() error {
	if cmd.dryRun {
		cmd.newLogEntry(cmd.lastBlock).Warn("?????? Dry run. Masternode is still on!")
		return nil
	}

	go cmd.sendBotMessage("???? Setting masternode off...")

	tx := cmd.config.Minter.TransactionOff

	if len(cmd.config.Minter.Seeds) > 0 {
		t, err := cmd.minter.GenerateCandidateOffTransaction(
			cmd.config.Minter.PublicKey,
			cmd.controlAddress,
			cmd.config.Minter.Seeds...,
		)

		if err != nil {
			cmd.log.Errorf("failed to generate transaction: %s", err)
		} else {
			tx = t
		}
	}

	resp, err := cmd.minter.SendTransaction(tx)

	if err != nil {
		return err
	}

	if resp.Error != nil {
		cmd.newLogEntry(cmd.lastBlock).Errorln(resp.Error)
		return errors.New(fmt.Sprintf("[%d] %s", resp.Error.Code, resp.Error.Message))
	}

	return nil
}

func (cmd *Command) lastBlockHeight() (int, error) {
	status, err := cmd.minter.Status()

	if err != nil {
		return 0, err
	}

	return status.LatestBlockHeight, nil
}

func (cmd *Command) isSigned(height int) (bool, error) {
	block, err := cmd.minter.GetBlock(height)

	if err != nil {
		return false, err
	}

	if len(block.Validators) == 0 {
		return false, NoValidatorsSignedYet
	}

	for _, validator := range block.Validators {
		if validator.PublicKey == cmd.config.Minter.PublicKey && validator.Signed {
			return true, nil
		}
	}

	return false, nil
}

func (cmd *Command) cleanupMissedBlocks() {
	if len(cmd.missedBlocks) == 0 {
		return
	}

	var temp []int

	for _, height := range cmd.missedBlocks {
		if cmd.lastBlock-height < cmd.config.Minter.MissedBlockRemoveAfter {
			temp = append(temp, height)
		}
	}

	cmd.missedBlocks = temp

	if cmd.prometheus != nil {
		cmd.prometheus.SetBlocksMissedCurrent(len(cmd.missedBlocks))
	}
}

func (cmd *Command) newLogEntry(height int) *logrus.Entry {
	return cmd.log.WithField("height", height).WithField("missed", len(cmd.missedBlocks))
}
