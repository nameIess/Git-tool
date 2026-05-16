package gitcfg

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/nameIess/git-tool/internal/logger"
	"github.com/nameIess/git-tool/internal/runner"
)

func ReadCurrent() (name, email string) {
	nameRes := runner.Run("git", "config", "--global", "user.name")
	emailRes := runner.Run("git", "config", "--global", "user.email")

	return strings.TrimSpace(nameRes.Stdout), strings.TrimSpace(emailRes.Stdout)
}

func Apply(name, email string, global bool) error {
	scope := "--global"
	if !global {
		scope = "--local"
	}

	logger.Info("Setting git config (%s): name=%q email=%q", scope, name, email)

	if res := runner.Run("git", "config", scope, "user.name", name); !res.Success() {
		return fmt.Errorf("failed to set user.name: %s", res.CombinedOutput())
	}
	if res := runner.Run("git", "config", scope, "user.email", email); !res.Success() {
		return fmt.Errorf("failed to set user.email: %s", res.CombinedOutput())
	}
	return nil
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
