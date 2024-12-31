package linep

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/berquerant/linep/internal"
	"github.com/berquerant/structconfig"
)

type Config struct {
	InitCmd    [][]string `json:"initCmd" name:"init" usage:"override commands to modify script environment; separated by semicolon"`
	PreExecCmd []string   `json:"preCmd" name:"preCmd" usage:"additional command to modify script; separated by space; stdout will be modified file content"`
	ExecCmd    []string   `json:"cmd" name:"cmd" usage:"override command to execute script; separated by space"`
	Imports    []string   `json:"import" name:"import" short:"i" usage:"additional libraries; separated by pipe"`
	Template   string     `json:"tmpl" name:"tmpl" usage:"override script template or template filename"`
	Main       string     `json:"main" name:"main" usage:"override script filename"`
	Dry        bool       `json:"dry" name:"dry" usage:"do not run; display generated script"`
	Debug      bool       `json:"debug" name:"debug" usage:"enable debug logs"`
	Quiet      bool       `json:"quiet" name:"quiet" short:"q" usage:"quiet stderr logs"`
	Keep       bool       `json:"keep" name:"keep" usage:"keep generated script directory"`
	Lang       string     `json:"lang"`
	Init       string     `json:"init"`
	Map        string     `json:"map"`
	Reduce     string     `json:"reduce"`
	WorkDir    string     `json:"workDir" name:"workDir" short:"w" usage:"working directory" default:".linep"`
	PWD        string     `json:"pwd"`
}

func (c Config) SetupLogger() { internal.SetupLogger(c.Debug, c.Quiet) }

func (Config) unmarshalCallback(f structconfig.StructField, v string, fv func() reflect.Value) error {
	n, _ := f.Tag().Name()
	switch n {
	case "init":
		if v == "" {
			return nil
		}
		xss := [][]string{}
		for _, x := range strings.Split(v, ";") {
			xss = append(xss, strings.Split(x, " "))
		}
		fv().Set(reflect.ValueOf(xss))
		return nil
	case "import":
		if v == "" {
			return nil
		}
		xs := strings.Split(v, "|")
		fv().Set(reflect.ValueOf(xs))
		return nil
	case "cmd", "preCmd":
		if v == "" {
			return nil
		}
		xs := strings.Split(v, " ")
		fv().Set(reflect.ValueOf(xs))
		return nil
	default:
		return fmt.Errorf("unexpected field: %s=%s", n, v)
	}
}

func (Config) equalCallback(a, b any) (bool, error) {
	switch a := a.(type) {
	case []string:
		b, ok := b.([]string)
		if !ok {
			return false, nil
		}
		return slices.Equal(a, b), nil
	case [][]string:
		b, ok := b.([][]string)
		if !ok {
			return false, nil
		}
		return slices.EqualFunc(a, b, slices.Equal), nil
	default:
		return false, fmt.Errorf("equalCallback got unexpected type: %#v, %#v", a, b)
	}
}

func NewStructConfig() *structconfig.StructConfig[Config] {
	var c Config
	return structconfig.New[Config](structconfig.WithAnyCallback(c.unmarshalCallback))
}

func NewConfigMerger() *structconfig.Merger[Config] {
	var c Config
	return structconfig.NewMerger[Config](
		structconfig.WithAnyCallback(c.unmarshalCallback),
		structconfig.WithAnyEqual(c.equalCallback),
	)
}
