package linep

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func NewConfig(fs *pflag.FlagSet) (*Config, error) {
	var (
		flagConfig, envConfig Config
		sc                    = NewStructConfig()
		merger                = NewConfigMerger()
	)

	if err := sc.FromEnv(&envConfig); err != nil {
		return nil, err
	}
	if err := sc.SetFlags(fs); err != nil {
		return nil, err
	}
	if err := fs.Parse(os.Args); err != nil {
		return nil, err
	}
	if err := sc.FromFlags(&flagConfig, fs); err != nil {
		return nil, err
	}
	config, err := merger.Merge(envConfig, flagConfig)
	if err != nil {
		return nil, err
	}

	// positional arguments
	switch fs.NArg() {
	case 3: // LANG MAP
		config.Map = fs.Arg(2)
	case 4: // LANG INIT MAP
		config.Init = fs.Arg(2)
		config.Map = fs.Arg(3)
	case 5: // LANG INIT MAP REDUCE
		config.Init = fs.Arg(2)
		config.Map = fs.Arg(3)
		config.Reduce = fs.Arg(4)
	default:
		return nil, fmt.Errorf("require 1 - 4 positional arguments")
	}
	config.Lang = fs.Arg(1)

	return &config, nil
}
