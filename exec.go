package linep

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/berquerant/execx"
	"gopkg.in/yaml.v3"
)

type Executor struct {
	Shell           []string
	Template        *Template
	Args            *ScriptArgs
	ExecPWD         string
	WorkDir         string
	KeepScript      bool
	Dry             bool
	DisplayTemplate bool

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
	dir, err := MkdirTemp(e.WorkDir, "linep")
	if err != nil {
		return err
	}
	e.tmpDir = dir

	return nil
}

func (e *Executor) Execute(ctx context.Context) error {
	if e.DisplayTemplate {
		if err := e.displayTemplate(os.Stdout); err != nil {
			return fmt.Errorf("%w: display template", err)
		}
		return nil
	}

	if e.Dry {
		if err := e.dump(os.Stdout); err != nil {
			return fmt.Errorf("%w: dry run", err)
		}
		return nil
	}

	slog.Debug("init")
	if err := e.init(); err != nil {
		return fmt.Errorf("%w: exec init", err)
	}
	slog.Debug("render")
	if err := e.renderTemplate(); err != nil {
		return fmt.Errorf("%w: render template", err)
	}
	slog.Debug("run:init")
	// redirect init output to stderr
	if err := e.runScript(ctx, nil, e.Stderr, e.Stderr, e.Template.Init); err != nil {
		return fmt.Errorf("%w: run init", err)
	}
	slog.Debug("run:exec")
	if err := e.runScript(ctx, e.Stdin, e.Stdout, e.Stderr, e.Template.Exec); err != nil {
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

func (e Executor) displayTemplate(w io.Writer) error {
	b, err := yaml.Marshal(e.Template)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s", b)
	return err
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
		"SRC_DIR",
		"WORK_DIR",
	}
	var (
		v = make([]string, len(seed)*2)
		i int
	)
	for _, x := range seed {
		v[i] = "@" + x
		i++
		v[i] = fmt.Sprintf(`"${%s}"`, x)
		i++
	}
	return strings.NewReplacer(v...).Replace(s)
}

func (e Executor) newEnv() execx.Env {
	env := execx.EnvFromEnviron()
	env.Set("EXEC_PWD", e.ExecPWD)
	env.Set("MAIN", e.Template.Main)
	env.Set("SRC_DIR", filepath.Dir(e.scriptFilename()))
	env.Set("WORK_DIR", e.WorkDir)
	return env
}

func (e Executor) runScript(
	ctx context.Context,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	script string,
) error {
	script = e.replaceMacros(script)
	s := execx.NewScript(script, e.Shell[0], e.Shell[1:]...)
	s.Env = e.newEnv()
	// s.KeepScriptFile = e.KeepScript

	logAttr := []any{
		slog.String("dir", e.tmpDir),
		slog.String("script", script),
		slog.String("expaned_script", s.Env.Expand(script)),
	}
	slog.Debug("executor:run", logAttr...)

	err := s.Runner(func(cmd *execx.Cmd) error {
		cmd.Dir = e.tmpDir
		cmd.Stdin = stdin
		_, err := cmd.Run(
			ctx,
			execx.WithStdoutWriter(new(execx.NullBuffer)),
			execx.WithStderrWriter(new(execx.NullBuffer)),
			execx.WithStdoutConsumer(e.logConsumer(stdout)),
			execx.WithStderrConsumer(e.logConsumer(stderr)),
		)
		return err
	})
	slog.Debug("executor:end", append(logAttr, WithErr(err))...)
	return err
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
	slog.Debug("executor:close", slog.String("work_dir", e.WorkDir))
	return os.RemoveAll(e.WorkDir)
}
