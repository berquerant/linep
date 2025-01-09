package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/berquerant/linep"
	"github.com/spf13/pflag"
)

func failOnError(err error) {
	if err != nil {
		slog.Error("exit", "err", fmt.Sprintf("%v", err))
		os.Exit(1)
	}
}

func main() {
	fs := pflag.NewFlagSet("main", pflag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, "linep")
		fs.PrintDefaults()
	}

	config, err := linep.NewConfig(fs)
	if errors.Is(err, pflag.ErrHelp) {
		return
	}
	failOnError(err)
	config.SetupLogger()
	slog.Debug("config", "body", fmt.Sprintf("%#v", config), "pflag.args", fmt.Sprintf("%#v", fs.Args()))

	if err := func() error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		e, err := config.Executor(os.Stdin, os.Stdout, os.Stderr)
		if err != nil {
			return err
		}
		return e.Execute(ctx)
	}(); err != nil {
		failOnError(err)
	}
}

const usage = `%[1]s -- process lines by one liner

Usage:
%[1]s LANG MAP [FLAGS]
%[1]s LANG INIT MAP [FLAGS]
%[1]s LANG INIT MAP REDUCE [FLAGS]

LANG: go, py, python, pipenv, rs, rust.

Requirements:
- go
- python, pipenv, pyenv
- cargo

Examples:
> seq 3 | %[1]s go 'fmt.Println(x+"0")'
10
20
30

> seq 10 | %[1]s rust 'let n:i32=x.parse().unwrap();if n%%2==0{println!("{}",n)}'
2
4
6
8
10

> seq 4 | %[1]s pipenv 'acc=[]' 'acc.append(int(x));print(math.prod(acc))' 'print(sum(acc))' --import 'math'
1
2
6
24
10

# without pipenv
> seq 3 | %[1]s pipenv 'print(x+"0")' --cmd python --init 'sleep 0'
10
20
30
# is almost equivalent to
> seq 3 | %[1]s py 'print(x+"0")'

# display formatted script via preCmd
> %[1]s pipenv 'acc=[]' 'acc.append(int(x));print(math.prod(acc))' 'print(sum(acc))' --import 'math' --preCmd 'black --quiet' --dry
import sys
import signal
import math

signal.signal(signal.SIGPIPE, signal.SIG_DFL)
acc = []
try:
    for x in sys.stdin:
        x = x.rstrip()
        acc.append(int(x))
        print(math.prod(acc))
except BrokenPipeError:
    pass
print(sum(acc))

# indent MAP (python)
> %[1]s py 'r={}' 'x=x.split(".")[-1]
if x in r:
  r[x]+=1
else:
  r[x]=1' 'for k, v in r.items():
  print(f"{k}\t{v}")' --dry
import sys
import signal
signal.signal(signal.SIGPIPE, signal.SIG_DFL)
r={}
try:
  for x in sys.stdin:
    x = x.rstrip()
    x=x.split(".")[-1]
    if x in r:
      r[x]+=1
    else:
      r[x]=1
except BrokenPipeError:
  pass
for k, v in r.items():
  print(f"{k}\t{v}")

Environment variables:
You can use the flag name with the hyphen removed and converted to uppercase as an environment variable.
If both the corresponding flag and the environment variable are specified at the same time, the flag takes precedence.

PWD is the temporary directory where the generated script exists.
If you want to refer to the directory from which linep is executed, please use EXEC_PWD.

Flags:
`
