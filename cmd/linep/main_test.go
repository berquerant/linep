package main_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	e := newExecutor(t)
	defer e.close()

	if err := run(os.Stdout, nil, e.cmd, "-h"); err != nil {
		t.Fatalf("%s help %v", e.cmd, err)
	}

	for _, tc := range []struct {
		title string
		input string
		args  []string
		want  string
	}{
		{
			title: "rust map",
			input: `1
2
3`,
			args: []string{
				"rust",
				`println!("{}0", x);`,
			},
			want: `10
20
30
`,
		},
		{
			title: "go map",
			input: `1
2
3`,
			args: []string{
				"go",
				`fmt.Println(x+"0")`,
			},
			want: `10
20
30
`,
		},
		{
			title: "py map",
			input: `1
2
3`,
			args: []string{
				"py",
				`print(x+"0")`,
			},
			want: `10
20
30
`,
		},
		{
			title: "go init map reduce",
			input: `1
2
3
4`,
			args: []string{
				"go",
				`acc := []int{};prod:=func()int{v:=1;for _, x := range acc {v*=x};return v};sum:=func()int{v:=0;for _, x := range acc {v+=x};return v}`,
				`i, _ := strconv.Atoi(x);acc=append(acc,i);fmt.Println(prod())`,
				`fmt.Println(sum())`,
				"--import",
				"strconv",
			},
			want: `1
2
6
24
10
`,
		},
		{
			title: "py init map reduce",
			input: `1
2
3
4`,
			args: []string{
				"pipenv",
				`acc=[]`,
				`acc.append(int(x));print(math.prod(acc))`,
				`print(sum(acc))`,
				"--import",
				"math",
			},
			want: `1
2
6
24
10
`,
		},
		{
			title: "bash",
			input: `1
2
3`,
			args: []string{
				"bash",
				`awk '{print $1 + 1}'`,
				"--cmd", "bash",
				"--tmpl", `#!/bin/bash
{{.Map}}`,
				"--main", "main.sh",
			},
			want: `2
3
4
`,
		},
		{
			title: "py map without pipenv",
			input: `1
2
3`,
			args: []string{
				"pipenv",
				`print(x+"0")`,
				"--cmd", "python",
				"--init", "sleep 0",
			},
			want: `10
20
30
`,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			t.Logf("run: %v", tc.args)
			stdin := bytes.NewBufferString(tc.input)
			var stdout bytes.Buffer

			args := []string{"--workDir", t.TempDir()}
			args = append(args, tc.args...)
			assert.Nil(t, run(&stdout, stdin, e.cmd, args...))
			assert.Equal(t, tc.want, stdout.String())
		})
	}
}

func run(w io.Writer, r io.Reader, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type executor struct {
	dir string
	cmd string
}

func newExecutor(t *testing.T) *executor {
	t.Helper()
	e := &executor{}
	e.init(t)
	return e
}

func (e *executor) init(t *testing.T) {
	t.Helper()
	dir, err := os.MkdirTemp("", "linep")
	if err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(dir, "linep")
	// build command
	if err := run(os.Stdout, nil, "go", "build", "-o", cmd); err != nil {
		t.Fatal(err)
	}
	e.dir = dir
	e.cmd = cmd
}

func (e *executor) close() {
	os.RemoveAll(e.dir)
}
