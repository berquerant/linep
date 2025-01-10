package linep

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func Stderr(quiet bool) io.Writer {
	if quiet {
		return io.Discard
	}
	return os.Stderr
}

func SetupLogger(debug, quiet bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(Stderr(quiet), &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

func WithErr(err error) any {
	return slog.String("err", fmt.Sprintf("%v", err))
}
