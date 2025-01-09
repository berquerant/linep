package linep

import (
	"errors"
	"fmt"
	"io"
	"text/template"

	"github.com/Masterminds/sprig"
)

var (
	ErrInvalidTemplate = errors.New("InvalidTemplate")
)

type Template struct {
	Name   string   `json:"name" yaml:"name"`
	Alias  []string `json:"alias" yaml:"alias"`
	Script string   `json:"script" yaml:"script"`
	Init   string   `json:"init" yaml:"init"`
	Exec   string   `json:"exec" yaml:"exec"`
	Main   string   `json:"main" yaml:"main"`
}

func (t *Template) Override(
	script, init, exec, main string,
) {
	if script != "" {
		t.Script = script
	}
	if init != "" {
		t.Init = init
	}
	if exec != "" {
		t.Exec = exec
	}
	if main != "" {
		t.Main = main
	}
}

func (t Template) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("%w: no name", ErrInvalidTemplate)
	}
	if t.Main == "" {
		return fmt.Errorf("%w: no main", ErrInvalidTemplate)
	}
	return nil
}

type ScriptArgs struct {
	Init   string   `json:"init" yaml:"init"`
	Map    string   `json:"map" yaml:"map"`
	Reduce string   `json:"reduce" yaml:"reduce"`
	Import []string `json:"import" yaml:"import"`
}

func (t Template) Execute(w io.Writer, args *ScriptArgs) error {
	x, err := template.New(t.Name).Funcs(sprig.FuncMap()).Parse(t.Script)
	if err != nil {
		return err
	}
	return x.Execute(w, args)
}
