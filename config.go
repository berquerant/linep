package linep

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/berquerant/structconfig"
	"gopkg.in/yaml.v3"
)

var (
	errUnexpectedType  = errors.New("UnexpectedType")
	errUnexpectedField = errors.New("UnexpectedField")
)

type Config struct {
	Dry             bool     `json:"dry" yaml:"dry" name:"dry" usage:"do not run; display generated script"`
	Debug           bool     `json:"debug" yaml:"debug" name:"debug" usage:"enable debug logs"`
	Quiet           bool     `json:"quiet" yaml:"quiet" name:"quiet" short:"q" usage:"quiet stderr logs"`
	Keep            bool     `json:"keep" yaml:"keep" name:"keep" usage:"keep generated script directory"`
	WorkDir         string   `json:"workDir" yaml:"workDir" name:"workDir" short:"w" usage:"working directory; default: $HOME/.linep"`
	Shell           []string `json:"sh" yaml:"sh" name:"sh" default:"sh" usage:"execute shell command; separated by ';'"`
	TemplateName    string   `json:"name" yaml:"name"`
	TemplateScript  string   `json:"tscript" yaml:"tscript" name:"script" usage:"override script"`
	TemplateInit    string   `json:"tinit" yaml:"tinit" name:"init" usage:"override init script"`
	TemplateExec    string   `json:"texec" yaml:"texec" name:"exec" usage:"override exec script"`
	TemplateMain    string   `json:"tmain" yaml:"tmain" name:"main" usage:"override main script name"`
	Init            string   `json:"init" yaml:"init"`
	Map             string   `json:"map" yaml:"map"`
	Reduce          string   `json:"reduce" yaml:"reduce"`
	Import          []string `json:"import" yaml:"import" name:"import" short:"i" usage:"additional libraries; separated by '|'"`
	PWD             string   `json:"pwd" yaml:"pwd"`
	DisplayTemplate bool     `json:"displayTemplate" yaml:"displayTemplate" name:"displayTemplate" usage:"do not run; display template"`
}

func (c *Config) Initialize() error {
	if c.WorkDir == "" {
		x, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		c.WorkDir = filepath.Join(x, ".linep")
	}

	return nil
}

func (c Config) Executor(stdin io.Reader, stdout io.Writer) (*Executor, error) {
	t, err := c.Template()
	if err != nil {
		return nil, err
	}
	return &Executor{
		Shell:           c.Shell,
		Template:        t,
		Args:            c.SciprtArgs(),
		ExecPWD:         c.PWD,
		WorkDir:         c.WorkDir,
		KeepScript:      c.Keep,
		Dry:             c.Dry,
		DisplayTemplate: c.DisplayTemplate,
		Stdin:           stdin,
		Stdout:          stdout,
		Stderr:          Stderr(c.Quiet),
	}, nil
}

func (c Config) Template() (*Template, error) {
	t, err := c.selectTemplate()
	if err != nil {
		return nil, err
	}
	t.Override(
		c.TemplateScript,
		c.TemplateInit,
		c.TemplateExec,
		c.TemplateMain,
	)
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func (c Config) selectTemplate() (*Template, error) {
	x, ok := builtinTemplates.get(c.TemplateName)
	if !ok {
		x, err := c.loadTemplate()
		if err != nil {
			return nil, fmt.Errorf("%w: load template %s", err, c.TemplateName)
		}
		return x, nil
	}
	return x, nil
}

func (c Config) loadTemplate() (*Template, error) {
	f, err := os.Open(c.TemplateName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var t Template
	if err := yaml.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c Config) SciprtArgs() *ScriptArgs {
	return &ScriptArgs{
		Init:   c.Init,
		Map:    c.Map,
		Reduce: c.Reduce,
		Import: c.Import,
	}
}

func (c Config) SetupLogger() {
	SetupLogger(c.Debug, c.Quiet)
}

func (Config) equalCallback(a, b any) (bool, error) {
	switch a := a.(type) {
	case []string:
		b, ok := b.([]string)
		if !ok {
			return false, nil
		}
		return slices.Equal(a, b), nil
	default:
		return false, fmt.Errorf("%w: %#v, %#v", errUnexpectedType, a, b)
	}
}

func (c Config) unmarshalCallback(f structconfig.StructField, v string, fv func() reflect.Value) error {
	name, ok := f.Tag().Name()
	if !ok {
		return nil
	}
	switch name {
	case "sh", "alias":
		if v == "" {
			return nil
		}
		fv().Set(reflect.ValueOf(strings.Split(v, ";")))
		return nil
	case "import":
		if v == "" {
			return nil
		}
		fv().Set(reflect.ValueOf(strings.Split(v, "|")))
		return nil
	default:
		return fmt.Errorf("%w: %s=%s", errUnexpectedField, name, v)
	}
}

func (c Config) StructConfig() *structconfig.StructConfig[Config] {
	return structconfig.New[Config](
		structconfig.WithAnyCallback(c.unmarshalCallback),
	)
}

func (c Config) Merger() *structconfig.Merger[Config] {
	return structconfig.NewMerger[Config](
		structconfig.WithAnyCallback(c.unmarshalCallback),
		structconfig.WithAnyEqual(c.equalCallback),
	)
}
