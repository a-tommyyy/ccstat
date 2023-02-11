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
	hashKey      = "HASH"
	hashFmt      = hashKey + ":%H"
	treeKey      = "TREE"
	treeFmt      = treeKey + ":%T"
	authorKey    = "AUTHOR"
	authorFmt    = authorKey + ":%an"
	committerKey = "COMMITTER"
	committerFmt = committerKey + ":%cn"
	subjectKey   = "SUBJECT"
	subjectFmt   = subjectKey + ":%s"
	bodyKey      = "BODY"
	bodyFmt      = bodyKey + ":%b"
	separator    = "@@__GIT_LOG_SEPARATOR__@@"
	delimiter    = "@@__GIT_LOG_DELIMITER__@@"

	logFmt = separator +
		hashFmt + delimiter +
		treeFmt + delimiter +
		authorFmt + delimiter +
		committerFmt + delimiter +
		subjectFmt + delimiter +
		bodyFmt
	prettyFmt = "--pretty=\"" + logFmt + "\""
)

type Commit struct {
	Hash      string
	Tree      string
	Author    string
	Committer string
	Subject   string
	Body      string
	Stat      *CommitStat
}

type ConventionalCommit struct {
	Commit *Commit
	Scope  string
	Type   string
}

type CommitStat struct {
	Insertion int
	Deletion  int
}

type GitConfig struct {
	Bin string
}

type GitClient interface {
	CanExec() error
	IsInsideWorkTree() error
	Exec(string, ...string) (string, error)
	Logs() ([]*Commit, error)
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

func (client *gitClientImpl) Logs() ([]*Commit, error) {
	args := []string{
		prettyFmt,
		"--no-decorate",
		"--no-merges",
		"--shortstat",
	}
	res, err := client.Exec("log", args...)
	if err != nil {
		return nil, err
	}
	return client.parseCommits(res)
}

func (client *gitClientImpl) parseCommits(logs string) ([]*Commit, error) {
	lines := strings.Split(logs, separator)[1:]

	commits := make([]*Commit, len(lines))
	for i, line := range lines {
		commits[i] = client.parseCommit(&line)
	}
	return commits, nil
}

func (client *gitClientImpl) parseCommit(log *string) *Commit {
	segments := strings.Split(*log, delimiter)
	commit := &Commit{}

	for _, segment := range segments {
		endFieldIdx := strings.Index(segment, ":")
		field := segment[0:endFieldIdx]
		content := segment[endFieldIdx+1:]

		switch field {
		case hashKey:
			commit.Hash = content
		case treeKey:
			commit.Tree = content
		case authorKey:
			commit.Author = content
		case committerKey:
			commit.Committer = content
		case subjectKey:
			commit.Subject = content
		case bodyKey:
			commit.Body = client.parseCommitBody(content)
		}
	}
	fmt.Println(commit.Body)
	return commit
}

func (client *gitClientImpl) parseCommitBody(body string) string {
	newLineDelimiter := "\n"
	body = strings.NewReplacer(
		"\r\n",
		newLineDelimiter,
		"\r",
		newLineDelimiter,
	).Replace(body)

	//TODO: Slice stat segment

	body = strings.TrimSpace(body)
	body = strings.Trim(body, "\"")
	body = strings.TrimSpace(body)
	body = strings.Trim(body, "\"")
	body = strings.TrimSpace(body)
	return body
}
