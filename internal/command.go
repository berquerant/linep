package internal

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func NewCmd(v ...string) Cmd {
	return Cmd(v)
}

type Cmd []string

func (c Cmd) Len() int         { return len(c) }
func (c Cmd) Arg(v string) Cmd { return Cmd(append(c, v)) }

//go:generate go run github.com/berquerant/goconfig -field "StdinReader io.Reader|StdoutWriter io.Writer|StderrWriter io.Writer|Dir string" -option -output command_config_generated.go -configOption Option

func (c Cmd) Execute(ctx context.Context, opt ...Option) error {
	if len(c) == 0 {
		return nil
	}

	config := NewConfigBuilder().
		StdinReader(nil).
		StdoutWriter(nil).
		StderrWriter(os.Stderr).
		Dir(".").
		Build()
	config.Apply(opt...)

	slog.Debug("run.execute",
		"cmd", strings.Join(c, " "),
		"dir", config.Dir.Get(),
	)
	x := exec.CommandContext(ctx, c[0], c[1:]...)
	x.Env = os.Environ()
	x.Dir = config.Dir.Get()
	x.Stdin = config.StdinReader.Get()
	x.Stdout = config.StdoutWriter.Get()
	x.Stderr = config.StderrWriter.Get()

	return x.Run()
}
