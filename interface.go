package linep

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func NewConfig(fs *pflag.FlagSet) (*Config, error) {
	var (
		defaultConfig Config
		err           error
		sc            = defaultConfig.StructConfig()
		merger        = defaultConfig.Merger()
	)
	if err := sc.FromDefault(&defaultConfig); err != nil {
		return nil, err
	}
	var envConfig Config
	if err := sc.FromEnv(&envConfig); err != nil {
		return nil, err
	}
	if envConfig, err = merger.Merge(defaultConfig, envConfig); err != nil {
		return nil, err
	}
	if err := sc.SetFlags(fs); err != nil {
		return nil, err
	}
	if err := fs.Parse(os.Args); err != nil {
		return nil, err
	}
	var flagConfig Config
	if err := sc.FromFlags(&flagConfig, fs); err != nil {
		return nil, err
	}
	if flagConfig, err = merger.Merge(defaultConfig, flagConfig); err != nil {
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
		if !config.DisplayTemplate {
			return nil, fmt.Errorf(
				"require 1 - 4 positional arguments: args: %v positional: %v",
				os.Args, fs.Args(),
			)
		}
	}
	config.TemplateName = fs.Arg(1)

	if x, err := os.Getwd(); err == nil {
		config.PWD = x
	} else {
		config.PWD = "."
	}

	c := &config
	if err := c.Initialize(); err != nil {
		return nil, err
	}

	return c, nil
}
