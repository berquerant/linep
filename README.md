# linep

``` shell
❯ linep
linep -- process lines by one liner

Usage:
linep LANG MAP [FLAGS]
linep LANG INIT MAP [FLAGS]
linep LANG INIT MAP REDUCE [FLAGS]

LANG: go, py, python, pipenv, rs, rust.

Requirements:
- go
- python, pipenv, pyenv
- cargo

Examples:
> seq 3 | linep go 'fmt.Println(x+"0")'
10
20
30

> seq 10 | linep rust 'let n:i32=x.parse().unwrap();if n%2==0{println!("{}",n)}'
2
4
6
8
10

> seq 4 | linep pipenv 'acc=[]' 'acc.append(int(x));print(math.prod(acc))' 'print(sum(acc))' --import 'math'
1
2
6
24
10

# without pipenv
> seq 3 | linep pipenv 'print(x+"0")' --cmd python --init 'sleep 0'
10
20
30
# is almost equivalent to
> seq 3 | linep py 'print(x+"0")'

# display formatted script via preCmd
> linep pipenv 'acc=[]' 'acc.append(int(x));print(math.prod(acc))' 'print(sum(acc))' --import 'math' --preCmd 'black --quiet' --dry
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
> linep py 'r={}' 'x=x.split(".")[-1]
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
      --cmd string       override command to execute script; separated by space
      --debug            enable debug logs
      --dry              do not run; display generated script
      --import string    additional libraries; separated by pipe
      --init string      override commands to modify script environment; separated by semicolon
      --keep             keep generated script directory
      --main string      override script filename
      --preCmd string    additional command to modify script; separated by space; stdout will be modified file content
      --quiet            quiet stderr logs
      --tmpl string      override script template or template filename
      --workDir string   working directory (default ".linep")
```
