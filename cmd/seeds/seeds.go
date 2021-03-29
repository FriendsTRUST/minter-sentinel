package seeds

import (
	"fmt"
	"minter-sentinel/config"
	"minter-sentinel/services/minter/node"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

type Command struct {
	log    *logrus.Logger
	config *config.Config

	minter *node.Service
}

func New(log *logrus.Logger, config *config.Config) *Command {
	return &Command{
		log:    log,
		config: config,
	}
}

func (cmd *Command) Command() *cli.Command {
	return &cli.Command{
		Name:  "seeds",
		Usage: "Get seed(s) of wallet(s)",
		Action: func(ctx *cli.Context) error {
			if svc, err := node.New(cmd.config.Minter.NodeApi, cmd.config.Minter.Testnet, cmd.log); err != nil {
				return err
			} else {
				cmd.minter = svc
			}

			return cmd.run()
		},
	}
}

func (cmd *Command) run() error {
	fmt.Printf("Testnet: %t\n", cmd.config.Minter.Testnet)

	var seeds []string

	for {
		fmt.Printf("Seed phrase %d (hidden): ", len(seeds)+1)
		byteSeed, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		seed := string(byteSeed)

		if len(seed) == 0 {
			break
		}

		wallet, err := cmd.minter.Wallet(seed, "")

		if err != nil {
			return err
		}

		seeds = append(seeds, wallet.Seed)
	}

	fmt.Println("Seeds:")

	for _, seed := range seeds {
		fmt.Println(seed)
	}

	return nil
}
