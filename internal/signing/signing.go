package signing

import (
	"fmt"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/runner"
)

type SigningStatus struct {
	Format string
	SigningKey string
	CommitSign string
	TagSign string
}

type configEntry struct {
	key string
	value string
}

func Configure(keyPath string, global bool) error {
	scope := "--global"
	if !global { scope = "--local" }

	pubKeyPath := keyPath + ".pub"
	logger.Info("Configuring Git for SSH commit signing (key: %s)", pubKeyPath)

	entries := []configEntry{
		{key:"gpg.format", value:"ssh"},
		{key:"user.signingkey", value:pubKeyPath},
		{key:"commit.gpgsign", value:"true"},
		{key:"tag.gpgsign", value:"true"},
	}

	type previousValue struct { key, value string; exists bool }
	previous := make([]previousValue, 0, len(entries))
	for _, entry := range entries {
		res := runner.Run("git", "config", scope, "--get", entry.key)
		if res.Success() {
			previous = append(previous, previousValue{key:entry.key,value:strings.TrimSpace(res.Stdout),exists:true})
		} else {
			previous = append(previous, previousValue{key:entry.key,exists:false})
		}
	}

	for _, entry := range entries {
		res := runner.Run("git", "config", scope, entry.key, entry.value)
		if !res.Success() {
			logger.Warn("Failed to set %s: %s; rolling back signing changes", entry.key, res.CombinedOutput())
			for _, old := range previous {
				if old.exists {
					_ = runner.Run("git", "config", scope, old.key, old.value)
				} else {
					_ = runner.Run("git", "config", scope, "--unset", old.key)
				}
			}
			return fmt.Errorf("failed to set %s", entry.key)
		}
	}
	logger.Info("SSH commit signing configured successfully")
	return nil
}

func Verify() (SigningStatus, error) {
	var status SigningStatus
	status.Format = strings.TrimSpace(runner.Run("git", "config", "gpg.format").Stdout)
	status.SigningKey = strings.TrimSpace(runner.Run("git", "config", "user.signingkey").Stdout)
	status.CommitSign = strings.TrimSpace(runner.Run("git", "config", "commit.gpgsign").Stdout)
	status.TagSign = strings.TrimSpace(runner.Run("git", "config", "tag.gpgsign").Stdout)
	return status, nil
}
