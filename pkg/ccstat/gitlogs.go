package ccstat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/atomiyama/ccstat/pkg/gitcmd"
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
	statKey      = "STAT"
	separator    = "@@__GIT_LOG_SEPARATOR__@@"
	delimiter    = "@@__GIT_LOG_DELIMITER__@@"
	logFmt       = separator +
		hashFmt + delimiter +
		treeFmt + delimiter +
		authorFmt + delimiter +
		committerFmt + delimiter +
		subjectFmt + delimiter +
		bodyFmt + delimiter +
		statKey + ":"
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

type GitLogs interface {
	Logs(*Options) ([]*ConventionalCommit, error)
}

type gitLogsImpl struct {
	git gitcmd.Git
}

type Options struct {
	After      string
	Before     string
	FollowPath string
}

func (gitlogs *gitLogsImpl) Logs(opt *Options) ([]*ConventionalCommit, error) {
	args := gitlogs.buildLogsArgs(opt)
	res, err := gitlogs.git.Exec("log", args...)
	if err != nil {
		return nil, err
	}
	return gitlogs.parseCommits(res)
}

func (gitlogs *gitLogsImpl) buildLogsArgs(opt *Options) []string {
	args := []string{
		prettyFmt,
		"--no-decorate",
		"--no-merges",
		"--shortstat",
	}
	if opt != nil {
		if opt.After != "" {
			args = append(args, fmt.Sprintf("--after=%s", opt.After))
		}
		if opt.Before != "" {
			args = append(args, fmt.Sprintf("--before=%s", opt.Before))
		}
		if opt.FollowPath != "" {
			args = append(args, opt.FollowPath)
		}
	}
	return args
}

func (gitlogs *gitLogsImpl) parseCommits(logs string) ([]*ConventionalCommit, error) {
	lines := strings.Split(logs, separator)[1:]

	commits := make([]*ConventionalCommit, len(lines))
	for i, line := range lines {
		commits[i] = gitlogs.parseConventionalCommit(line)
	}
	return commits, nil
}

func (gitlogs *gitLogsImpl) parseConventionalCommit(log string) *ConventionalCommit {
	commit := gitlogs.parseCommit(log)
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

func (gitlogs *gitLogsImpl) parseCommit(log string) *Commit {
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
			commit.Body = gitlogs.parseCommitBody(content)
		case statKey:
			commit.Stat = gitlogs.parseCommitStat(content)
		}
	}
	return commit
}

func (gitlogs *gitLogsImpl) parseCommitBody(body string) string {
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

func (gitlogs *gitLogsImpl) parseCommitStat(body string) *CommitStat {
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
