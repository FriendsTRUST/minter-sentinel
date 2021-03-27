package config

import (
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Config struct {
	Telegram   Telegram   `yaml:"telegram"`
	Minter     Minter     `yaml:"minter"`
	Prometheus Prometheus `yaml:"prometheus"`
}
type Telegram struct {
	Token  string `yaml:"token"`
	Admins []int  `yaml:"admins"`
}

type Minter struct {
	Testnet                bool   `yaml:"testnet"`
	ApiUrl                 string `yaml:"api_url"`
	ExplorerApiUrl         string `yaml:"explorer_api_url"`
	GateApiUrl             string `yaml:"gate_api_url"`
	PublicKey              string `yaml:"public_key"`
	TransactionOff         string `yaml:"transaction_off"`
	MissedBlocksThreshold  int    `yaml:"missed_blocks_threshold"`
	Sleep                  int    `yaml:"sleep"`
	MissedBlockRemoveAfter int    `yaml:"missed_block_remove_after"`
}

type Prometheus struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

func New(path string) (*Config, error) {
	var cfg Config

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: true,
		SkipFiles:    false,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{path},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	if err := loader.Load(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
