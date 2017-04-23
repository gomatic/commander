package commander

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

//
type Executor func(string, []string, []string) error

//
func Launch(prefix string, args []string, env []string, exe func(string, []string, []string) error) error {

	if len(args) == 0 {
		return errors.New("Missing required command parameter.")
	}

	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	return exe(binary, args, env)
}

//
func Exec(binary string, args []string, env []string) error {
	return syscall.Exec(binary, args, env)
}

//
func Spawn(binary string, args []string, env []string) error {
	cmd := exec.Command(binary, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Env = env
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}
