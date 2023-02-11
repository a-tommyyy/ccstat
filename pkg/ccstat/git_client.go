package ccstat

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	/*
	 * git log format constants for git client pretty format argument
	 */
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
	logFmt       = separator +
		hashFmt + delimiter +
		treeFmt + delimiter +
		authorFmt + delimiter +
		committerFmt + delimiter +
		subjectFmt + delimiter +
		bodyFmt
	prettyFmt = "--pretty=\"" + logFmt + "\""
)

var (
	/*
	 * Regexp for git log --shortstat
	 */
	statRe = regexp.MustCompile(
		`(?P<file>[\d]*) files? changed(?:, (?P<insertion>[\d]*) insertions?\(\+\))?(?:, (?P<deletion>[\d]*) deletions?\(\-\))?`,
	)

	/*
	 * Regexp for Conventional Commit Header
	 */
	headerRe = regexp.MustCompile(`^(?P<type>\w*)(?:\((?P<scope>.*)\))?: (?P<subject>.*)$`)
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
	Type      string
	Scope     string
	Subject   string
	Body      string
	RawCommit *Commit
}

type CommitStat struct {
	FileChanged int
	Insertion   int
	Deletion    int
}

type GitConfig struct {
	GitBin string
}

type GitClient interface {
	Logs() ([]*ConventionalCommit, error)
	IsInsideWorkTree() error
}

type gitClientImpl struct {
	config *GitConfig
}

func (client *gitClientImpl) Logs() ([]*ConventionalCommit, error) {
	args := []string{
		prettyFmt,
		"--no-decorate",
		"--no-merges",
		"--shortstat",
	}
	res, err := client.exec("log", args...)
	if err != nil {
		return nil, err
	}
	return client.parseCommits(res)
}

func newGitClient(config *GitConfig) GitClient {
	bin := "git"
	if config != nil && config.GitBin != "" {
		bin = config.GitBin
	}

	return &gitClientImpl{
		config: &GitConfig{
			GitBin: bin,
		},
	}
}

func (client *gitClientImpl) canExec() error {
	_, err := exec.LookPath(client.config.GitBin)
	if err != nil {
		return fmt.Errorf("\"%s\" not found", client.config.GitBin)
	}
	return nil
}

func (client *gitClientImpl) IsInsideWorkTree() error {
	out, err := client.exec("rev-parse", "--is-inside-work-tree")
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

func (client *gitClientImpl) exec(subcmd string, args ...string) (string, error) {
	// Check git command available
	if err := client.canExec(); err != nil {
		return "", err
	}

	// Build git sub-commands
	commands := append([]string{subcmd}, args...)

	// Run command
	var out bytes.Buffer
	cmd := exec.Command(client.config.GitBin, commands...)
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

func (client *gitClientImpl) parseCommits(logs string) ([]*ConventionalCommit, error) {
	lines := strings.Split(logs, separator)[1:]

	commits := make([]*ConventionalCommit, len(lines))
	for i, line := range lines {
		commits[i] = client.parseConventionalCommit(line)
	}
	return commits, nil
}

func (client *gitClientImpl) parseConventionalCommit(log string) *ConventionalCommit {
	commit := client.parseCommit(log)
	conventionalCommit := &ConventionalCommit{RawCommit: commit}
	if headerRe.MatchString(commit.Subject) {
		match := headerRe.FindStringSubmatch(commit.Subject)
		for i, name := range headerRe.SubexpNames() {
			if i != 0 && name != "" {
				switch name {
				case "type":
					conventionalCommit.Type = match[i]
				case "scope":
					conventionalCommit.Scope = match[i]
				case "subject":
					conventionalCommit.Subject = match[i]
				}
			}

		}
	}
	return conventionalCommit
}

func (client *gitClientImpl) parseCommit(log string) *Commit {
	segments := strings.Split(log, delimiter)
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
			commit.Stat = client.parseCommitStat(content)
		}
	}
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

	body = strings.TrimSpace(body)
	body = strings.Trim(body, "\"")
	body = strings.TrimSpace(body)
	body = strings.Trim(body, "\"")
	body = strings.TrimSpace(body)
	return body
}

func (client *gitClientImpl) parseCommitStat(body string) *CommitStat {
	stat := &CommitStat{}
	if !statRe.MatchString(body) {
		return &CommitStat{}
	}

	match := statRe.FindStringSubmatch(body)
	for i, name := range statRe.SubexpNames() {
		if i != 0 && name != "" {
			number, err := strconv.Atoi(match[i])
			if err != nil {
				continue
			}
			switch name {
			case "insertion":
				stat.Insertion = number
			case "deletion":
				stat.Deletion = number
			case "file":
				stat.FileChanged = number
			}
		}
	}
	return stat
}
