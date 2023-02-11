package ccstat

import "fmt"

type StatRow struct {
	Index     string
	Insertion int
	Deletion  int
	SumOfDiff int
}

func AggByScope() (string, error) {
	client := NewGitClient(&GitConfig{})

	commits, err := client.Logs()
	if err != nil {
		return "", err
	}
	result := make(map[string]*StatRow)
	for _, commit := range commits {
		index := commit.Scope
		if index == "" {
			index = "None"
		}
		if result[index] == nil {
			result[index] = &StatRow{Index: commit.Scope}
		}
		result[index].Insertion += commit.RawCommit.Stat.Insertion
		result[index].Deletion += commit.RawCommit.Stat.Deletion
		result[index].SumOfDiff = result[index].Insertion + result[index].Deletion
	}
	for key, value := range result {
		fmt.Printf("SCOPE:%s\tINSERT:%v\tDELETE:%v\tSUM:%v\n", key, value.Insertion, value.Deletion, value.SumOfDiff)
	}
	return "", nil
}
