package sshkey

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/runner"
)

func Exists(keyPath string) bool {
	_, err := os.Stat(keyPath)
	return err == nil
}

func Generate(keyPath, email, passphrase string) error {
	logger.Info("Generating SSH key: type=ed25519 email=%s path=%s", email, keyPath)

	if err := platform.EnsureSSHDir(); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	args := []string{"-t", "ed25519", "-C", email, "-f", keyPath}
	if passphrase == "" {
		args = append(args, "-N", "")
	} else {
		args = append(args, "-N", passphrase)
	}

	res := runner.Run("ssh-keygen", args...)
	if !res.Success() {
		return fmt.Errorf("ssh-keygen failed: %s", res.CombinedOutput())
	}
	return nil
}

func ReadPublicKey(keyPath string) (string, error) {
	data, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func FindNextKeyName(basePath string) string {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	for i := 2; i <= 99; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d", base, i))
		if !Exists(candidate) {
			return candidate
		}
	}
	return basePath + "_new"
}

func WriteSSHConfig(keyPath string) error {
	configPath := filepath.Join(platform.SSHDir(), "config")
	entry := fmt.Sprintf(`
Host github.com
    IdentityFile %s
    AddKeysToAgent yes
`, keyPath)

	content, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if !strings.Contains(string(content), "Host github.com") {
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(entry)
		return err
	}
	return nil
}
