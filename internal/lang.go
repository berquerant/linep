package internal

import "path/filepath"

type Lang int

//go:generate go run golang.org/x/tools/cmd/stringer -type Lang -output lang_stringer_generated.go

const (
	Lunknown Lang = iota
	Lgo
	Lpython
	LpythonPipenv
	Lrust
)

func NewLang(s string) Lang {
	switch s {
	case "go":
		return Lgo
	case "py", "python":
		return Lpython
	case "pipenv":
		return LpythonPipenv
	case "rust", "rs":
		return Lrust
	default:
		return Lunknown
	}
}

func (la Lang) Spec() LangSpec {
	switch la {
	case Lgo:
		return &GoLangSpec{}
	case Lpython:
		return &PythonLangSpec{}
	case LpythonPipenv:
		var x PythonPipenvLangSpec
		return &x
	case Lrust:
		return &RustLangSpec{}
	default:
		return &EmptyLangSpec{}
	}
}

type LangSpec interface {
	Main() string
	InitCmd(dir string) [][]string
	ExecCmd() []string
	Template() Template
}

type GoLangSpec struct{}

func (GoLangSpec) Main() string { return "main.go" }
func (GoLangSpec) InitCmd(dir string) [][]string {
	return [][]string{
		{"go", "mod", "init", filepath.Base(dir)},
		{"go", "mod", "tidy"},
		{"go", "fmt"},
	}
}
func (GoLangSpec) ExecCmd() []string  { return []string{"go", "run"} }
func (GoLangSpec) Template() Template { return NewGoTemplate() }

type PythonLangSpec struct{}

func (PythonLangSpec) Main() string                { return "main.py" }
func (PythonLangSpec) InitCmd(_ string) [][]string { return nil }
func (PythonLangSpec) ExecCmd() []string           { return []string{"python"} }
func (PythonLangSpec) Template() Template          { return NewPythonTemplate() }

type PythonPipenvLangSpec struct {
	PythonLangSpec
}

func (PythonPipenvLangSpec) InitCmd(_ string) [][]string {
	return [][]string{
		{"pipenv", "install"},
	}
}
func (PythonPipenvLangSpec) ExecCmd() []string { return []string{"pipenv", "run", "python"} }

type RustLangSpec struct{}

func (RustLangSpec) Main() string { return "main.rs" }
func (s RustLangSpec) InitCmd(_ string) [][]string {
	return [][]string{
		{"cargo", "init"},
		{"cargo", "update"},
	}
}
func (RustLangSpec) ExecCmd() []string  { return []string{"cargo", "run"} }
func (RustLangSpec) Template() Template { return NewRustTemplate() }

type EmptyLangSpec struct{}

func (EmptyLangSpec) Main() string                { return "" }
func (EmptyLangSpec) InitCmd(_ string) [][]string { return nil }
func (EmptyLangSpec) ExecCmd() []string           { return nil }
func (EmptyLangSpec) Template() Template          { return NewEmptyTemplate() }
