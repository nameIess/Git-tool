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

const (
	managedConfigStart = "# BEGIN Git-Tool managed github.com"
	managedConfigEnd = "# END Git-Tool managed github.com"
)

func Exists(keyPath string) bool { _, err := os.Stat(keyPath); return err == nil }

func Generate(keyPath, email, passphrase string) error {
	logger.Info("Generating SSH key: type=ed25519 email=%s path=%s", email, keyPath)
	if err := platform.EnsureSSHDir(); err != nil { return fmt.Errorf("failed to create .ssh directory: %w", err) }
	if Exists(keyPath) || Exists(keyPath+".pub") {
		if !Exists(keyPath) || !Exists(keyPath+".pub") { return fmt.Errorf("incomplete existing SSH key pair at %s", keyPath) }
		return generateReplacement(keyPath, email, passphrase)
	}
	return generateToPath(keyPath, email, passphrase)
}

func generateToPath(keyPath, email, passphrase string) error {
	args := []string{"-t", "ed25519", "-C", email, "-f", keyPath, "-N", passphrase}
	res := runner.RunRedacted("ssh-keygen", []int{7}, args...)
	if !res.Success() { return fmt.Errorf("ssh-keygen failed: %s", res.CombinedOutput()) }
	return ensureKeyPair(keyPath)
}

func generateReplacement(keyPath, email, passphrase string) error {
	dir := filepath.Dir(keyPath)
	base := filepath.Base(keyPath)
	tempPath := filepath.Join(dir, "."+base+".git-tool-new")
	_ = os.Remove(tempPath); _ = os.Remove(tempPath+".pub")
	if err := generateToPath(tempPath, email, passphrase); err != nil { cleanupGenerated(tempPath); return err }

	backupPath := filepath.Join(dir, "."+base+".git-tool-backup")
	backupPubPath := backupPath+".pub"
	_ = os.Remove(backupPath); _ = os.Remove(backupPubPath)
	if err := os.Rename(keyPath, backupPath); err != nil { cleanupGenerated(tempPath); return fmt.Errorf("failed to stage existing private key: %w", err) }
	if err := os.Rename(keyPath+".pub", backupPubPath); err != nil {
		_ = os.Rename(backupPath, keyPath); cleanupGenerated(tempPath)
		return fmt.Errorf("failed to stage existing public key: %w", err)
	}
	if err := os.Rename(tempPath, keyPath); err != nil {
		_ = os.Rename(backupPubPath, keyPath+".pub"); _ = os.Rename(backupPath, keyPath); cleanupGenerated(tempPath)
		return fmt.Errorf("failed to install new private key: %w", err)
	}
	if err := os.Rename(tempPath+".pub", keyPath+".pub"); err != nil {
		_ = os.Remove(keyPath); _ = os.Rename(backupPubPath, keyPath+".pub"); _ = os.Rename(backupPath, keyPath); cleanupGenerated(tempPath)
		return fmt.Errorf("failed to install new public key: %w", err)
	}
	_ = os.Remove(backupPath); _ = os.Remove(backupPubPath)
	return ensureKeyPair(keyPath)
}

func cleanupGenerated(path string) { _ = os.Remove(path); _ = os.Remove(path+".pub") }
func ensureKeyPair(keyPath string) error {
	if !Exists(keyPath) || !Exists(keyPath+".pub") { return fmt.Errorf("SSH key generation did not produce the expected key pair") }
	return nil
}

func ReadPublicKey(keyPath string) (string, error) {
	data, err := os.ReadFile(keyPath+".pub")
	if err != nil { return "", err }
	return strings.TrimSpace(string(data)), nil
}

func FindNextKeyName(basePath string) string {
	dir := filepath.Dir(basePath); base := filepath.Base(basePath)
	for i:=2; i<=99; i++ {
		candidate:=filepath.Join(dir,fmt.Sprintf("%s_%d",base,i))
		if !Exists(candidate) && !Exists(candidate+".pub") { return candidate }
	}
	return basePath+"_new"
}

func WriteSSHConfig(keyPath string) error {
	configPath:=filepath.Join(platform.SSHDir(),"config")
	data,err:=os.ReadFile(configPath)
	if err!=nil && !os.IsNotExist(err){return err}
	content:=string(data)
	managed:=fmt.Sprintf("%s\nHost github.com\n    IdentityFile %s\n    IdentitiesOnly yes\n    AddKeysToAgent yes\n%s",managedConfigStart,keyPath,managedConfigEnd)

	if start:=strings.Index(content,managedConfigStart);start>=0 {
		if end:=strings.Index(content[start:],managedConfigEnd);end>=0 {
			end+=start+len(managedConfigEnd)
			content=strings.TrimRight(content[:start],"\r\n")+"\n"+managed+content[end:]
			return writeConfig(configPath,content)
		}
	}

	lines:=strings.Split(strings.ReplaceAll(content,"\r\n","\n"),"\n")
	hostStart,hostEnd:=-1,len(lines)
	for i,line:=range lines {
		trimmed:=strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed),"host ") {
			fields:=strings.Fields(trimmed)
			if len(fields)==2 && strings.EqualFold(fields[1],"github.com") {
				hostStart=i
				for j:=i+1;j<len(lines);j++ {
					next:=strings.TrimSpace(lines[j])
					if strings.HasPrefix(strings.ToLower(next),"host ")||strings.HasPrefix(strings.ToLower(next),"match "){hostEnd=j;break}
				}
				break
			}
		}
	}
	if hostStart>=0 {
		block:=append([]string{},lines[hostStart:hostEnd]...)
		filtered:=[]string{block[0]}
		for _,line:=range block[1:] {
			lower:=strings.ToLower(strings.TrimSpace(line))
			if strings.HasPrefix(lower,"identityfile ")||strings.HasPrefix(lower,"identitiesonly ")||strings.HasPrefix(lower,"addkeystoagent "){continue}
			filtered=append(filtered,line)
		}
		filtered=append(filtered,"    IdentityFile "+keyPath,"    IdentitiesOnly yes","    AddKeysToAgent yes")
		lines=append(append(append([]string{},lines[:hostStart]...),filtered...),lines[hostEnd:]...)
		return writeConfig(configPath,strings.TrimRight(strings.Join(lines,"\n"),"\n")+"\n")
	}
	trimmed:=strings.TrimRight(content,"\r\n")
	if trimmed==""{return writeConfig(configPath,managed+"\n")}
	return writeConfig(configPath,trimmed+"\n\n"+managed+"\n")
}

func writeConfig(path,content string)error{
	if err:=platform.EnsureSSHDir();err!=nil{return err}
	return os.WriteFile(path,[]byte(content),0600)
}
