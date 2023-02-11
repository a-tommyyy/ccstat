package ccstat

import (
	"fmt"
	"os"
	"path/filepath"
)

type Row struct {
	Index     string
	Insertion int
	Deletion  int
	SumOfDiff int
}

type Result struct {
	GroupBy string
	After   string
	Before  string
	Rows    []Row
}

type CCStat struct {
	gitClient GitClient
	config    *Config
}

type Config struct {
	RepoPath string
	GitBin   string
}

func New(cnf *Config) *CCStat {
	if cnf == nil {
		cnf = &Config{}
	}
	return &CCStat{
		config:    cnf,
		gitClient: newGitClient(&GitConfig{GitBin: cnf.GitBin}),
	}
}

func (ccs *CCStat) AggByScope() (string, error) {
	back, err := ccs.workingDir()
	if err != nil {
		return "", err
	}
	defer back()
	if err := ccs.gitClient.IsInsideWorkTree(); err != nil {
		return "", err
	}

	commits, err := ccs.gitClient.Logs()
	if err != nil {
		return "", err
	}
	result := make(map[string]*Row)
	for _, commit := range commits {
		index := commit.Scope
		if index == "" {
			index = "None"
		}
		ccs.aggregate(index, result, commit)
	}
	for key, value := range result {
		fmt.Printf("SCOPE:%s\tINSERT:%v\tDELETE:%v\tSUM:%v\n", key, value.Insertion, value.Deletion, value.SumOfDiff)
	}
	return "", nil
}

func (ccs *CCStat) workingDir() (func() error, error) {
	current, err := filepath.Abs(".")
	back := func() error {
		return os.Chdir(current)
	}
	if err != nil {
		return back, err
	}

	repoPath, err := filepath.Abs(ccs.config.RepoPath)
	if err != nil {
		return back, err
	}
	if err := os.Chdir(repoPath); err != nil {
		return back, err
	}
	return back, nil
}

func (ccs *CCStat) aggregate(idx string, accum map[string]*Row, commit *ConventionalCommit) map[string]*Row {
	if accum[idx] == nil {
		accum[idx] = &Row{Index: idx}
	}
	result := accum[idx]
	stat := commit.RawCommit.Stat
	result.Insertion += stat.Insertion
	result.Deletion += stat.Deletion
	result.SumOfDiff = stat.Insertion + stat.Deletion
	return accum
}
