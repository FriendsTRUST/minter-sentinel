package txgenerate

import (
	"bufio"
	"errors"
	"fmt"
	"minter-sentinel/config"
	"minter-sentinel/services/minter/node"
	"os"
	"strings"
	"syscall"

	"github.com/MinterTeam/minter-go-sdk/v2/wallet"
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
		Name:  "txgenerate",
		Usage: "Generate transaction to turn off masternode",
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
	fmt.Printf("Public Key: %s", cmd.config.Minter.PublicKey)
	fmt.Println()

	var wallets []*wallet.Wallet

	for {
		fmt.Printf("Seed phrase %d (hidden, leave blank to finish): ", len(wallets)+1)
		byteMnemonic, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		mnemonic := string(byteMnemonic)

		if len(mnemonic) == 0 {
			break
		}

		wal, err := cmd.minter.Wallet(mnemonic, "")

		if err != nil {
			return err
		}

		wallets = append(wallets, wal)
	}

	if len(wallets) == 0 {
		return errors.New("enter at least 1 seed phrase")
	}

	var walletAddress string

	if len(wallets) > 1 {
		fmt.Printf("Multisig address: ")

		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		walletAddress = text
	} else {
		walletAddress = wallets[0].Address
	}

	var seeds []string

	for _, wal := range wallets {
		seeds = append(seeds, wal.Seed)
	}

	tx, err := cmd.minter.GenerateCandidateOffTransaction(
		cmd.config.Minter.PublicKey,
		walletAddress,
		seeds...,
	)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to generate transaction: %s", err))
	}

	fmt.Printf("Transaction:\n%s", tx)
	fmt.Println()

	return nil
}
