package commander

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

//
type Commanding struct {
	Cmd       *exec.Cmd
	Prefix    string
	Binary    string
	debugging bool
}

//
func New(prefix string) *Commanding {
	debugging, exists := os.LookupEnv("DEBUGGING")
	return &Commanding{
		Prefix:    prefix,
		Cmd:       &exec.Cmd{},
		debugging: exists && strings.ToLower(debugging) == "true",
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
	args := os.Args[1:]
	la := len(args)
	if la == 0 || from > la {
		return c
	}
	start, end := from, la
	if c.debugging {
		log.Printf("from:%d start:%d end:%d la:%d args:%s", from, start, end, la, args)
	}
	if start < 0 {
		start = la + start
	}
	if end < 0 {
		end = la + end
	}
	if start > end {
		start, end = end, start
	}
	if start > la || start < 0 || end < 0 {
		return c
	} else if end > la {
		end = la
	}
	c.Cmd.Args = append(c.Cmd.Args, args[start:end]...)
	if c.debugging {
		log.Printf("from:%d start:%d end:%d la:%d args:%s", from, start, end, la, c.Cmd.Args)
	}
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
func (c Commanding) Execute() error {
	if c.Binary == "" {
		return fmt.Errorf("No binary specified")
	}

	if c.debugging {
		if config, err := yaml.Marshal(yaml.MapSlice{
			{"binary", c.Binary},
			{"args", c.Cmd.Args},
		}); err == nil {
			fmt.Fprintln(os.Stderr, string(config))
		}
	}

	return Launch(append([]string{c.Binary}, c.Cmd.Args...), c.Cmd.Env, Exec)
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
