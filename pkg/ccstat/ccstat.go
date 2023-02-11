package ccstat

import "fmt"

func AggByScope() (string, error) {
	client := NewGitClient(&GitConfig{})

	commits, err := client.Logs()
	if err != nil {
		return "", err
	}
	for _, c := range commits {
		fmt.Printf(
			"RAWHEADER:%s\nTYPE:%s\nSCOPE:%s\nSUBJECT:%s\nINSERTION:%v\nDELETION:%v\n",
			c.RawCommit.Subject,
			c.Type,
			c.Scope,
			c.Subject,
			c.RawCommit.Stat.Insertion,
			c.RawCommit.Stat.Deletion,
		)
		fmt.Println("----END COMMIT-----")
	}
	return "", nil
}
