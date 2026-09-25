package sshkey

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nameIess/git-tool/internal/platform"
)

func TestFindNextKeyNameSkipsPrivateAndPublicPairs(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(base, []byte("key"), 0600); err != nil { t.Fatal(err) }
	if err := os.WriteFile(base+"_2.pub", []byte("key"), 0644); err != nil { t.Fatal(err) }
	got := FindNextKeyName(base)
	want := filepath.Join(dir, "id_ed25519_3")
	if got != want { t.Fatalf("got %q, want %q", got, want) }
}

func TestWriteSSHConfigCreatesManagedBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	keyPath := filepath.Join(platform.SSHDir(), "id_ed25519")
	if err := WriteSSHConfig(keyPath); err != nil { t.Fatal(err) }
	data, err := os.ReadFile(filepath.Join(platform.SSHDir(), "config"))
	if err != nil { t.Fatal(err) }
	content := string(data)
	for _, want := range []string{managedConfigStart, "Host github.com", "IdentityFile "+keyPath, "IdentitiesOnly yes", "AddKeysToAgent yes", managedConfigEnd} {
		if !strings.Contains(content, want) { t.Fatalf("config missing %q:\n%s", want, content) }
	}
}

func TestWriteSSHConfigUpdatesExistingGithubHost(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	sshDir := platform.SSHDir()
	if err := os.MkdirAll(sshDir, 0700); err != nil { t.Fatal(err) }
	configPath := filepath.Join(sshDir, "config")
	initial := "Host github.com\n    IdentityFile ~/.ssh/old_key\n    AddKeysToAgent no\n\nHost example.com\n    IdentityFile ~/.ssh/example\n"
	if err := os.WriteFile(configPath, []byte(initial), 0600); err != nil { t.Fatal(err) }
	newKey := filepath.Join(sshDir, "new_key")
	if err := WriteSSHConfig(newKey); err != nil { t.Fatal(err) }
	data, err := os.ReadFile(configPath)
	if err != nil { t.Fatal(err) }
	content := string(data)
	if strings.Contains(content, "old_key") || strings.Contains(content, "AddKeysToAgent no") { t.Fatalf("old settings remained:\n%s", content) }
	if !strings.Contains(content, "IdentityFile "+newKey) { t.Fatalf("new key missing:\n%s", content) }
	if !strings.Contains(content, "Host example.com") { t.Fatalf("unrelated host was modified:\n%s", content) }
}
