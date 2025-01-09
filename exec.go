package linep

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/berquerant/execx"
)

type Executor struct {
	Shell      []string
	Template   *Template
	Args       *ScriptArgs
	ExecPWD    string
	WorkDir    string
	KeepScript bool
	Dry        bool

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	tmpDir string
}

func (e *Executor) init() error {
	if w := e.WorkDir; w != "" {
		if err := os.MkdirAll(w, 0755); err != nil {
			return err
		}
	}
	dir, err := os.MkdirTemp(e.WorkDir, "linep")
	if err != nil {
		return err
	}
	e.tmpDir = dir

	return nil
}

func (e *Executor) Execute(ctx context.Context) error {
	if e.Dry {
		if err := e.dump(os.Stdout); err != nil {
			return fmt.Errorf("%w: dry run", err)
		}
		return nil
	}

	if err := e.init(); err != nil {
		return fmt.Errorf("%w: exec init", err)
	}
	if err := e.renderTemplate(); err != nil {
		return fmt.Errorf("%w: render template", err)
	}
	if err := e.runScript(ctx, nil, e.Template.Init); err != nil {
		return fmt.Errorf("%w: run init", err)
	}
	if err := e.runScript(ctx, e.Stdin, e.Template.Exec); err != nil {
		return fmt.Errorf("%w: run exec", err)
	}

	return nil
}

func (e Executor) renderTemplate() error {
	f, err := os.Create(e.scriptFilename())
	if err != nil {
		return err
	}
	defer f.Close()

	return e.dump(f)
}

func (e Executor) dump(w io.Writer) error {
	var b bytes.Buffer
	if err := e.Template.Execute(&b, e.Args); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "%s", e.replaceMacros(b.String()))
	return err
}

func (e Executor) scriptFilename() string {
	return filepath.Join(e.tmpDir, e.Template.Main)
}

func (Executor) replaceMacros(s string) string {
	seed := []string{
		"EXEC_PWD",
		"MAIN",
		"MAIN_DIR",
		"WORK_DIR",
	}
	v := make([]string, len(seed)*2)
	for i, x := range seed {
		v[i] = "@" + x
		v[i+1] = fmt.Sprintf(`"%s"`, x)
	}
	return strings.NewReplacer(v...).Replace(s)
}

func (e Executor) newEnv() execx.Env {
	env := execx.EnvFromEnviron()
	env.Set("EXEC_PWD", e.ExecPWD)
	env.Set("MAIN", e.Template.Main)
	env.Set("MAIN_DIR", filepath.Dir(e.scriptFilename()))
	env.Set("WORK_DIR", e.WorkDir)
	return env
}

func (e Executor) runScript(
	ctx context.Context,
	stdin io.Reader,
	script string,
) error {
	s := execx.NewScript(script, e.Shell[0], e.Shell[1:]...)
	s.Env = e.newEnv()
	s.KeepScriptFile = e.KeepScript

	return s.Runner(func(cmd *execx.Cmd) error {
		cmd.Dir = e.tmpDir
		cmd.Stdin = stdin
		_, err := cmd.Run(
			ctx,
			execx.WithStdoutWriter(new(execx.NullBuffer)),
			execx.WithStderrWriter(new(execx.NullBuffer)),
			execx.WithStdoutConsumer(e.logConsumer(e.Stdout)),
			execx.WithStderrConsumer(e.logConsumer(e.Stderr)),
		)
		return err
	})
}

func (Executor) logConsumer(w io.Writer) func(execx.Token) {
	return func(t execx.Token) {
		fmt.Fprintln(w, t.String())
	}
}

func (e Executor) Close() error {
	if e.KeepScript {
		return nil
	}
	return os.RemoveAll(e.WorkDir)
}
