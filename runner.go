package linep

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/berquerant/linep/internal"
)

type Runner Config

func (r Runner) Run(ctx context.Context) error {
	var (
		attrs   []any
		addAttr = func(k string, v any) {
			attrs = append(attrs, k, v)
		}
	)

	slog.Debug("run", attrs...)
	if w := r.WorkDir; w != "" {
		if err := os.MkdirAll(w, 0755); err != nil {
			return err
		}
	}
	dir, err := os.MkdirTemp(r.WorkDir, "linep")
	if err != nil {
		return err
	}
	if !r.Keep {
		defer os.RemoveAll(dir)
	}
	addAttr("dir", dir)

	slog.Debug("template", attrs...)
	filename, err := r.newTemplate(dir)
	if err != nil {
		return err
	}
	addAttr("filename", filename)

	slog.Debug("init", attrs...)
	if err := r.runInit(ctx, dir); err != nil {
		return err
	}

	slog.Debug("pre.exec", attrs...)
	if err := r.runPreExec(ctx, filename); err != nil {
		return err
	}

	if r.Dry {
		return r.dumpScript(filename)
	}

	slog.Debug("exec", attrs...)
	if err := r.runExec(ctx, filename); err != nil {
		return err
	}
	return nil
}

func (r Runner) runExec(ctx context.Context, filename string) error {
	if len(r.execCmd()) == 0 {
		return nil
	}

	cc := internal.NewCmd(r.execCmd()...).Arg(filepath.Base(filename))
	return cc.Execute(
		ctx,
		internal.WithStdinReader(os.Stdin),
		internal.WithStdoutWriter(os.Stdout),
		internal.WithStderrWriter(r.stderr()),
		internal.WithDir(filepath.Dir(filename)),
	)
}

func (Runner) dumpScript(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(os.Stdout, f)
	return err
}

func (r Runner) runPreExec(ctx context.Context, filename string) error {
	if len(r.preExecCmd()) == 0 {
		return nil
	}

	cc := internal.NewCmd(r.preExecCmd()...).Arg(filepath.Base(filename))
	var buf bytes.Buffer
	if err := cc.Execute(
		ctx,
		internal.WithStdoutWriter(&buf), // stdout of preCmd will be script
		internal.WithStderrWriter(r.stderr()),
		internal.WithDir(filepath.Dir(filename)),
	); err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, &buf)
	return err
}

func (r Runner) runInit(ctx context.Context, dir string) error {
	for _, x := range r.initCmd(dir) {
		cc := internal.NewCmd(x...)
		if err := cc.Execute(
			ctx,
			internal.WithDir(dir),
			internal.WithStdoutWriter(r.stderr()), // redirect to stderr
			internal.WithStderrWriter(r.stderr()),
		); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) newTemplate(dir string) (string, error) {
	m, ok := r.main()
	if !ok {
		return "", fmt.Errorf("main is empty")
	}

	f, err := os.Create(filepath.Join(dir, m))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := r.renderTemplate(f); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (r Runner) renderTemplate(w io.Writer) error {
	t, err := r.template()
	if err != nil {
		return err
	}

	data := &internal.TemplateArgs{
		Init:    r.Init,
		Map:     r.Map,
		Reduce:  r.Reduce,
		Imports: r.Imports,
	}
	slog.Debug("render", "data", fmt.Sprintf("%#v", data))
	return t.Execute(w, data)
}

func (r Runner) initCmd(dir string) [][]string {
	x, _ := internal.NotEmpty(r.InitCmd, r.spec().InitCmd(dir))
	return x
}

func (r Runner) preExecCmd() []string { return r.PreExecCmd }

func (r Runner) execCmd() []string {
	x, _ := internal.NotEmpty(r.ExecCmd, r.spec().ExecCmd())
	return x
}

func (r Runner) template() (internal.Template, error) {
	if x := r.Template; x != "" {
		return internal.ParseTemplateStringOrFile(x)
	}
	return r.spec().Template(), nil
}

func (r Runner) main() (string, bool)    { return internal.NotEmpty(r.Main, r.spec().Main()) }
func (r Runner) spec() internal.LangSpec { return r.lang().Spec() }
func (r Runner) lang() internal.Lang     { return internal.NewLang(r.Lang) }
func (r Runner) stderr() io.Writer       { return internal.Stderr(r.Quiet) }
