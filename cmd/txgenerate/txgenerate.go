package txgenerate

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
	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:     "multisig",
			Required: false,
			Value:    false,
		},
	}

	return &cli.Command{
		Name:  "txgenerate",
		Usage: "Generate transaction to turn off masternode",
		Action: func(ctx *cli.Context) error {
			if svc, err := node.New(cmd.config.Minter.NodeApi, cmd.config.Minter.Testnet, cmd.log); err != nil {
				return err
			} else {
				cmd.minter = svc
			}

			if ctx.Bool("multisig") {
				return cmd.runMultisig()
			}

			return cmd.run()
		},
		Flags: flags,
	}
}

func (cmd *Command) run() error {
	fmt.Printf("Testnet: %t\n", cmd.config.Minter.Testnet)
	fmt.Printf("Public Key: %s", cmd.config.Minter.PublicKey)
	fmt.Println()
	fmt.Print("Seed phrase (hidden): ")
	byteSeed, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	fmt.Println()

	seed := string(byteSeed)

	wallet, err := cmd.minter.Wallet(seed)

	if err != nil {
		return err
	}

	tx, err := cmd.minter.GenerateCandidateOffTransaction(
		cmd.config.Minter.PublicKey,
		wallet,
		wallet.PrivateKey,
	)

	if err != nil {
		return err
	}

	fmt.Printf("Transaction:\n%s", tx)
	fmt.Println()

	return nil
}

func (cmd *Command) runMultisig() error {
	panic("implement")
}
