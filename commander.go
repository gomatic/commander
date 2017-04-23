package commander

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

//
type Commanding struct {
	Cmd    *exec.Cmd
	Prefix string
	Binary string
}

//
func New(prefix string) *Commanding {
	return &Commanding{
		Prefix: prefix,
		Cmd:    &exec.Cmd{},
	}
}

// Append args to the command arguments.
func (c *Commanding) Args(args ...string) *Commanding {
	c.Cmd.Args = append(c.Cmd.Args, args...)
	return c
}

// Append os.Args[from:] to the command arguments.
func (c *Commanding) Inherit(from int) *Commanding {
	c.Cmd.Env = os.Environ()
	if la := len(os.Args); from > la {
		from = la
	} else if from < 0 {
		from = 0
	}
	l := len(os.Args)
	if l == 0 {
		return c
	}
	start, end := from%l, l
	if start < 0 {
		start = l + start
	}
	if end < 0 {
		end = l + end
	}
	if start > end {
		start, end = end, start
	}
	if start > l || start < 0 || end < 0 {
		return c
	} else if end > l {
		end = l
	}
	c.Cmd.Args = append(c.Cmd.Args, os.Args[start:end]...)
	return c
}

// Append env to the command environment.
func (c *Commanding) Env(env ...string) *Commanding {
	c.Cmd.Env = append(c.Cmd.Env, env...)
	return c
}

// Append the map to the command environment as key=value for each key/value pair, quoting value.
func (c *Commanding) EnvMap(env map[string]interface{}) *Commanding {
	for n, v := range env {
		var s string
		switch v := v.(type) {
		case string:
			s = v
		default:
			s = strconv.QuoteToASCII(fmt.Sprintf("%+v", v))
			s = s[1 : len(s)-1]
		}
		c.Cmd.Env = append(c.Cmd.Env, fmt.Sprintf("%s=%s", n, s))
	}
	return c
}

// Looks up the prefix+name in the PATH. If found, assigns c.Binary.
// To execute a specific binary without first calling c.LookPath, just set c.Binary
func (c *Commanding) LookPath(name string) (*Commanding, error) {
	binary := c.Prefix + name
	binary, err := exec.LookPath(binary)
	if err != nil {
		return c, err
	}
	c.Binary = binary
	return c, nil
}

// Exec the command.
// To execute a specific binary without first calling c.LookPath, just set c.Binary
func (c *Commanding) Execute() error {
	if c.Binary == "" {
		return fmt.Errorf("%s", "")
	}
	return Launch(c.Prefix, append([]string{c.Binary}, c.Cmd.Args...), c.Cmd.Env, Exec)
}

// Reproduce a command line string that reflects a usable command line.
func (c *Commanding) String() string {

	quoter := func(e string) string {
		if !strings.Contains(e, " ") {
			return e
		}
		p := strings.SplitN(e, "=", 2)
		if strings.Contains(p[0], " ") {
			p[0] = `"` + strings.Replace(p[0], `"`, `\"`, -1) + `"`
		}
		if len(p) == 1 {
			return p[0]
		}
		return p[0] + `="` + strings.Replace(p[1], `"`, `\"`, -1) + `"`
	}
	each := func(s []string) (o []string) {
		o = make([]string, len(s))
		for i, t := range s {
			o[i] = quoter(t)
		}
		return
	}
	return c.Binary + " " + strings.Join(each(c.Cmd.Args), " ")
}
