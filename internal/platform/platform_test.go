package platform

import (
	"path/filepath"
	"testing"
)

func TestHomeAndSSHDir(t *testing.T) {
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	if HomeDir() == "" { t.Fatal("HomeDir returned empty path") }
	if filepath.Dir(SSHDir()) != HomeDir() { t.Fatalf("SSHDir is not under HomeDir: %q", SSHDir()) }
	if filepath.Base(DefaultKeyPath()) != "id_ed25519" { t.Fatalf("unexpected default key name: %q", DefaultKeyPath()) }
}
