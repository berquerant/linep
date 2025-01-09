package linep

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed template/empty.yml
var emptyTemplate []byte

//go:embed template/go.yml
var goTemplate []byte

//go:embed template/pipenv.yml
var pipenvTemplate []byte

//go:embed template/python.yml
var pythonTemplate []byte

//go:embed template/rust.yml
var rustTemplate []byte

type templateMap struct {
	d map[string]*Template
}

func (m *templateMap) add(t *Template) {
	m.d[t.Name] = t
	for _, a := range t.Alias {
		m.d[a] = t
	}
}

func (m templateMap) get(name string) (*Template, bool) {
	x, ok := m.d[name]
	return x, ok
}

func newTemplateMap() *templateMap {
	m := &templateMap{
		d: map[string]*Template{},
	}
	for _, b := range [][]byte{
		emptyTemplate,
		goTemplate,
		pipenvTemplate,
		pythonTemplate,
		rustTemplate,
	} {
		var x Template
		if err := yaml.Unmarshal(b, &x); err != nil {
			panic(err)
		}
		m.add(&x)
	}
	return m
}

var (
	builtinTemplates = newTemplateMap()
)
