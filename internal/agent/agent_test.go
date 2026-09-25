package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nameIess/git-tool/internal/platform"
)

func TestWriteBashrcSnippetUpdatesManagedBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	path := filepath.Join(home, ".bashrc")
	old := "# before\n" + bashrcMarker + "\nold\n" + bashrcEndMarker + "\n# after\n"
	if err := os.WriteFile(path, []byte(old), 0644); err != nil { t.Fatal(err) }
	keyPath := filepath.Join(platform.SSHDir(), "id_ed25519_2")
	if err := WriteBashrcSnippet(keyPath); err != nil { t.Fatal(err) }
	data, err := os.ReadFile(path)
	if err != nil { t.Fatal(err) }
	content := string(data)
	if !strings.Contains(content, "id_ed25519_2") { t.Fatalf("new key was not written:\n%s", content) }
	if strings.Contains(content, "\nold\n") { t.Fatalf("stale managed block remained:\n%s", content) }
	if !strings.Contains(content, "# before") || !strings.Contains(content, "# after") { t.Fatalf("unmanaged content was lost:\n%s", content) }
}
