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
		e, err := config.Executor(os.Stdin, os.Stdout)
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
%[1]s TEMPLATE MAP [FLAGS]
%[1]s TEMPLATE INIT MAP [FLAGS]
%[1]s TEMPLATE INIT MAP REDUCE [FLAGS]

TEMPLATE: go, py, python, pipenv, rs, rust, nil, null, empty

Requirements of templates:
go: go
python, py: python
pipenv: pipenv, pyenv
rust, rs: cargo

Examples:
> seq 3 | %[1]s go 'fmt.Println(x+"0")' -q
10
20
30

> seq 10 | %[1]s rust 'let n:i32=x.parse().unwrap();if n%%2==0{println!("{}",n)}' -q
2
4
6
8
10

> seq 4 | %[1]s pipenv 'acc=[]' 'acc.append(int(x));print(math.prod(acc))' 'print(sum(acc))' --import 'math' -q
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

Templates:
TEMPLATE argument can be a template filename.
empty (nil, null) template is for overriding.
A template file format is:

# template name, required.
name: sample
# generated script name, required.
main: main.go
# aliases of name.
# also used by TEMPLATE argument.
alias:
  - smpl
# template of script body (main.go, main.py, ...).
# executed by https://pkg.go.dev/text/template with https://masterminds.github.io/sprig/
# available fields:
#   Init   : INIT argument (string)
#   Map    : MAP argument (string)
#   Reduce : REDUCE argument (string)
#   Import : --import argument (slice of string)
script: |
  ...
# init script command.
# initialize a directory of generated script like 'go mod init'.
# macros are replaced with a reference of an environment variable.
# available macros:
#   @MAIN     : main of this template
#   @WORK_DIR : --workDir argument
#   @EXEC_PWD : current directory of %[1]s execution
#   @SRC_DIR  : directory of the generated script
init: |
  ...
# execute script command.
# execute generated script like 'go run @MAIN'.
# macros are available.
exec: |
  ...

# show template
> %[1]s go --displayTemplate

Environment variables:
You can use the flag name with the hyphen removed and converted to uppercase as an environment variable.
If both the corresponding flag and the environment variable are specified at the same time, the flag takes precedence.

Flags:
`
