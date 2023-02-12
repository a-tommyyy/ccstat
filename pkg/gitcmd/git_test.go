package gitcmd

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanExec(t *testing.T) {
	t.Run("With default gitbin", func(t *testing.T) {
		git := NewGit(nil)
		err := git.CanExec()
		assert.NoError(t, err)
	})

	t.Run("With local aware gitbin", func(t *testing.T) {
		bytes, _ := exec.Command("which", "git").Output()
		bin := strings.TrimSpace(string(bytes))
		config := &Config{GitBin: bin}
		git := NewGit(config)
		err := git.CanExec()
		assert.NoError(t, err)
	})

	t.Run("With invalid gitbin", func(t *testing.T) {
		config := &Config{GitBin: "/notfound/bin/git"}
		git := NewGit(config)
		err := git.CanExec()
		assert.Error(t, err)
	})
}
