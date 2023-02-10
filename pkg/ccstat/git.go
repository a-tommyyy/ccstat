package ccstat

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	hashFmt      = "HASH:%H"
	treeFmt      = "TREE:%T"
	authorFmt    = "AUTHOR:%an"
	committerFmt = "COMMITTER:%cn"
	subjectFmt   = "SUBJECT:%s"
	bodyFmt      = "BODY:%b"
	separator    = "@@__GIT_LOG_SEPARATOR__@@"
	delimiter    = "@@__GIT_LOG_DELIMITER__@@"

	logFmt = separator +
		hashFmt + delimiter +
		treeFmt + delimiter +
		authorFmt + delimiter +
		committerFmt + delimiter +
		subjectFmt + delimiter +
		bodyFmt
	prettyFmt = "--pretty\"" + logFmt + "\""
)

type Commit struct {
	Hash      string
	Tree      string
	Author    string
	Committer string
	Subject   string
	Body      string
}

type ConventionalCommit struct {
	Commit *Commit
	Scope  string
	Type   string
}

type GitConfig struct {
	Bin string
}

type GitClient interface {
	CanExec() error
	IsInsideWorkTree() error
	Exec(string, ...string) (string, error)
	Logs() ([]Commit, error)
}

type gitClientImpl struct {
	config *GitConfig
}

func NewGitClient(config *GitConfig) GitClient {
	bin := "git"
	if config != nil && config.Bin != "" {
		bin = config.Bin
	}

	return &gitClientImpl{
		config: &GitConfig{
			Bin: bin,
		},
	}
}

func (client *gitClientImpl) CanExec() error {
	_, err := exec.LookPath(client.config.Bin)
	if err != nil {
		return fmt.Errorf("\"%s\" not found", client.config.Bin)
	}
	return nil
}

func (client *gitClientImpl) IsInsideWorkTree() error {
	out, err := client.Exec("rev-parse", "--is-inside-work-tree")
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

func (client *gitClientImpl) Exec(subcmd string, args ...string) (string, error) {
	commands := append([]string{subcmd}, args...)

	var out bytes.Buffer
	cmd := exec.Command(client.config.Bin, commands...)
	cmd.Stdout = &out
	cmd.Stderr = io.Discard

	cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		return "", nil
	}

	return strings.TrimRight(strings.TrimSpace(out.String()), "\000"), nil
}

func (client *gitClientImpl) Logs() ([]Commit, error) {
	args := []string{
		prettyFmt,
		"--no-merges",
		"--shortstat",
		"-1", // For Testing
	}
	_, err := client.Exec("log", args...)
	if err != nil {
		return nil, err
	}
	return []Commit{}, nil
}
