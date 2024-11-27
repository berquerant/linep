package internal

import (
	_ "embed"
	"io"
	"os"
	"text/template"
)

type Template interface {
	Execute(w io.Writer, data *TemplateArgs) error
}

type TemplateArgs struct {
	Init    string
	Map     string
	Reduce  string
	Imports []string
}

type baseTemplate struct {
	tmpl *template.Template
}

func (t baseTemplate) Execute(w io.Writer, data *TemplateArgs) error {
	return t.tmpl.Execute(w, data)
}

func newBaseTemplate(name, text string) Template {
	return &baseTemplate{
		tmpl: template.Must(template.New(name).Parse(text)),
	}
}

//go:embed tmpl/go.tmpl
var goTemplate []byte

func NewGoTemplate() Template {
	return newBaseTemplate("go", string(goTemplate))
}

//go:embed tmpl/python.tmpl
var pythonTemplate []byte

func NewPythonTemplate() Template {
	return newBaseTemplate("python", string(pythonTemplate))
}

//go:embed tmpl/rust.tmpl
var rustTemplate []byte

func NewRustTemplate() Template {
	return newBaseTemplate("rust", string(rustTemplate))
}

type OtherTemplate struct {
	tmpl *template.Template
}

func (t OtherTemplate) Execute(w io.Writer, data *TemplateArgs) error {
	return t.tmpl.Execute(w, data)
}

func ParseTemplateStringOrFile(s string) (*OtherTemplate, error) {
	t, err := func() (*template.Template, error) {
		if f, err := os.Open(s); err == nil {
			defer f.Close()
			b, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}
			return template.New("other").Parse(string(b))
		}

		return template.New("other").Parse(s)
	}()

	if err != nil {
		return nil, err
	}
	return &OtherTemplate{
		tmpl: t,
	}, nil
}

func NewEmptyTemplate() *EmptyTemplate { return &EmptyTemplate{} }

type EmptyTemplate struct{}

func (EmptyTemplate) Execute(_ io.Writer, _ *TemplateArgs) error { return nil }
