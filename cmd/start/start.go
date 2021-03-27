package start

import (
	"errors"
	"fmt"
	"minter-sentinel/config"
	"minter-sentinel/services/minter/explorer"
	"minter-sentinel/services/minter/gate"
	"minter-sentinel/services/prometheus"
	"minter-sentinel/services/telegram"
	"sort"
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

	dryRun       bool
	missedBlocks []int
	lastBlock    int

	explorer   *explorer.Service
	gate       *gate.Service
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

			if !cmd.dryRun && len(cmd.config.Minter.TransactionOff) == 0 {
				return errors.New("`transaction_off` not set in configuration file")
			}

			if len(cmd.config.Minter.GateApiUrl) == 0 {
				return errors.New("define gate_api_url in configuration file")
			}

			if e, err := explorer.New(cmd.config.Minter.ExplorerApiUrl, cmd.log); err != nil {
				return err
			} else {
				cmd.explorer = e
			}

			if g, err := gate.New(cmd.config.Minter.GateApiUrl, cmd.log); err != nil {
				return err
			} else {
				cmd.gate = g
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

			lastBlock, err := cmd.lastBlockHeight()

			if err != nil {
				return err
			}

			signed, err := cmd.isSigned(lastBlock)

			if err != nil {
				return err
			}

			if !signed {
				return errors.New("last block is not signed by validator, start watcher when validator starts signing blocks")
			}

			cmd.lastBlock = lastBlock

			return cmd.run()
		},
	}
}

func (cmd *Command) run() error {
	cmd.newLogEntry(cmd.lastBlock).
		WithField("missed_blocks_threshold", cmd.config.Minter.MissedBlocksThreshold).
		WithField("sleep", cmd.config.Minter.Sleep).
		WithField("missed_block_remove_after", cmd.config.Minter.MissedBlockRemoveAfter).
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
					if _, ok := err.(*explorer.BlockNotFound); ok {
						cmd.newLogEntry(nextHeight).Debugln("Block not created yet.")
						continue
					}

					if errors.Is(err, NoValidatorsSignedYet) {
						continue
					}

					go cmd.sendBotMessage(fmt.Sprintf("⚠️ Failed to detect if block is signed: %s", err))

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

					go cmd.sendBotMessage(fmt.Sprintf("⚠️ Block %d missed [%d/%d]", nextHeight, len(cmd.missedBlocks), cmd.config.Minter.MissedBlocksThreshold))
				}

				if len(cmd.missedBlocks) >= cmd.config.Minter.MissedBlocksThreshold {
					cmd.newLogEntry(nextHeight).Errorln("Missed blocks threshold exceeded")

					go cmd.sendBotMessage(fmt.Sprintf("🚨 Block %d missed [%d/%d]", nextHeight, len(cmd.missedBlocks), cmd.config.Minter.MissedBlocksThreshold))

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
		go cmd.sendBotMessage("🚨 Sending transaction to turn off masternode")

		if err := cmd.turnOffMasternode(); err != nil {
			cmd.newLogEntry(cmd.lastBlock).Errorln("Failed to turn off masternode", err)

			cmd.sendBotMessage("🚨 Failed to turn off masternode")

			return err
		}

		go cmd.sendBotMessage("🚨 Masternode is off")
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
		cmd.newLogEntry(cmd.lastBlock).Warn("⚠️ Dry start. Masternode is still on!")
		return nil
	}

	go cmd.sendBotMessage("🚨 Setting masternode off...")

	resp, err := cmd.gate.SendTransaction(cmd.config.Minter.TransactionOff)

	if err != nil {
		return err
	}

	if resp.Error != nil {
		cmd.newLogEntry(cmd.lastBlock).Errorln(resp.Error)
		return errors.New(fmt.Sprintf("[%d] %s", resp.Error.Code, resp.Error.Log))
	}

	return nil
}

func (cmd *Command) lastBlockHeight() (int, error) {
	listBlocks, err := cmd.explorer.ListBlocks()

	if err != nil {
		return 0, err
	}

	items := make([]int, len(listBlocks.Data))

	for i, block := range listBlocks.Data {
		items[i] = block.Height
	}

	sort.Ints(items)

	return items[len(items)-1], nil
}

func (cmd *Command) isSigned(height int) (bool, error) {
	block, err := cmd.explorer.GetBlock(height)

	if err != nil {
		return false, err
	}

	if len(block.Data.Validators) == 0 {
		return false, NoValidatorsSignedYet
	}

	for _, validator := range block.Data.Validators {
		if validator.Validator.PublicKey == cmd.config.Minter.PublicKey && validator.Signed {
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

	lastHeight := cmd.missedBlocks[len(cmd.missedBlocks)-1]

	for _, height := range cmd.missedBlocks {
		if lastHeight-height >= cmd.config.Minter.MissedBlockRemoveAfter {
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
