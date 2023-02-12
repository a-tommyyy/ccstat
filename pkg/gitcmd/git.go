package gitcmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Git interface {
	CanExec() error
	IsInsideWorkTree() error
	Exec(string, ...string) (string, error)
}

type Config struct {
	GitBin string
}

type gitImpl struct {
	config *Config
}

func NewGit(cnf *Config) Git {
	bin := "git"
	if cnf != nil {
		if cnf.GitBin != "" {
			bin = cnf.GitBin
		}
	}
	return &gitImpl{
		config: &Config{
			GitBin: bin,
		},
	}
}

func (git *gitImpl) CanExec() error {
	_, err := exec.LookPath(git.config.GitBin)
	if err != nil {
		return fmt.Errorf("\"%s\" not found", git.config.GitBin)
	}
	return nil
}

func (git *gitImpl) Exec(subcmd string, args ...string) (string, error) {
	// Build git sub-commands
	commands := append([]string{subcmd}, args...)

	// Run command
	var out bytes.Buffer
	cmd := exec.Command(git.config.GitBin, commands...)
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	cmd.Run()

	// Handle command result
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		return "", nil
	}
	return strings.TrimRight(strings.TrimSpace(out.String()), "\000"), nil
}

func (git *gitImpl) IsInsideWorkTree() error {
	out, err := git.Exec("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return err
	}

	if out != "true" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		return fmt.Errorf("\"%s\" is not git repository", cwd)
	}
	return nil
}
