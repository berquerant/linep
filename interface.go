package linep

import (
	"fmt"
	"os"

	"github.com/berquerant/structconfig"
	"github.com/spf13/pflag"
)

func NewConfig(fs *pflag.FlagSet) (*Config, error) {
	var b Config
	config, err := structconfig.NewConfigWithMerge(b.StructConfig(), b.Merger(), fs)
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

	if err := config.Initialize(); err != nil {
		return nil, err
	}

	return config, nil
}
