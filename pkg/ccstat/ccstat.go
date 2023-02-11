package ccstat

func AggByScope() (string, error) {
	client := NewGitClient(&GitConfig{})

	_, err := client.Logs()
	if err != nil {
		return "", err
	}
	return "", nil
}
