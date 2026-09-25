package signing

import (
	"fmt"
	"strings"

	"github.com/nameIess/git-tool/internal/runner"
)

type SigningStatus struct{Format,SigningKey,CommitSign,TagSign string}
func Configure(keyPath string,global bool)error{scope:="--global";if !global{scope="--local"};cfg:=map[string]string{"gpg.format":"ssh","user.signingkey":keyPath+".pub","commit.gpgsign":"true","tag.gpgsign":"true"};for k,v:=range cfg{r:=runner.Run("git","config",scope,k,v);if !r.Success(){return fmt.Errorf("failed to set %s: %s",k,r.CombinedOutput())}};return nil}
func Verify()(SigningStatus,error){var s SigningStatus;s.Format=strings.TrimSpace(runner.Run("git","config","gpg.format").Stdout);s.SigningKey=strings.TrimSpace(runner.Run("git","config","user.signingkey").Stdout);s.CommitSign=strings.TrimSpace(runner.Run("git","config","commit.gpgsign").Stdout);s.TagSign=strings.TrimSpace(runner.Run("git","config","tag.gpgsign").Stdout);return s,nil}
