package ccstat

import "fmt"

func AggByScope() (string, error) {
	client := NewGitClient(&GitConfig{})

	res, err := client.Logs()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", res), nil
}
