package signing

import (
	"fmt"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/runner"
)

type SigningStatus struct {
	Format     string
	SigningKey string
	CommitSign string
	TagSign    string
}

func Configure(keyPath string, global bool) error {
	scope := "--global"
	if !global {
		scope = "--local"
	}

	pubKeyPath := keyPath + ".pub"
	logger.Info("Configuring Git for SSH commit signing (key: %s)", pubKeyPath)

	configs := map[string]string{
		"gpg.format":      "ssh",
		"user.signingkey": pubKeyPath,
		"commit.gpgsign":  "true",
		"tag.gpgsign":     "true",
	}

	for k, v := range configs {
		res := runner.Run("git", "config", scope, k, v)
		if !res.Success() {
			logger.Warn("Failed to set %s: %s", k, res.CombinedOutput())
			return fmt.Errorf("failed to set %s", k)
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
