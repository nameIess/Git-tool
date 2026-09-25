package sshkey

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nameIess/git-tool/internal/platform"
	"github.com/nameIess/git-tool/internal/runner"
)

func Exists(path string)bool{_,err:=os.Stat(path);return err==nil}
func Generate(path,email,passphrase string)error{if err:=platform.EnsureSSHDir();err!=nil{return fmt.Errorf("create SSH directory: %w",err)};args:=[]string{"-t","ed25519","-C",email,"-f",path,"-N",passphrase};if Exists(path){return fmt.Errorf("SSH key already exists: %s",path)};if r:=runner.RunRedacted("ssh-keygen",[]int{7},args...);!r.Success(){return fmt.Errorf("ssh-keygen failed: %s",r.CombinedOutput())};return nil}
func ReadPublicKey(path string)(string,error){b,err:=os.ReadFile(path+".pub");if err!=nil{return "",err};return strings.TrimSpace(string(b)),nil}
func FindNextKeyName(base string)string{dir,name:=filepath.Dir(base),filepath.Base(base);for i:=2;i<=99;i++{p:=filepath.Join(dir,fmt.Sprintf("%s_%d",name,i));if !Exists(p)&&!Exists(p+".pub"){return p}};return base+"_new"}
func WriteSSHConfig(keyPath string)error{if err:=platform.EnsureSSHDir();err!=nil{return err};path:=filepath.Join(platform.SSHDir(),"config");b,_:=os.ReadFile(path);content:=string(b);const begin="# BEGIN Git-Tool managed github.com";const end="# END Git-Tool managed github.com";block:=fmt.Sprintf("%s\nHost github.com\n    IdentityFile %s\n    IdentitiesOnly yes\n    AddKeysToAgent yes\n%s\n",begin,keyPath,end);if i:=strings.Index(content,begin);i>=0{if j:=strings.Index(content[i:],end);j>=0{j=i+j+len(end);content=content[:i]+block+content[j:];return os.WriteFile(path,[]byte(content),0600)}};if content!=""&&!strings.HasSuffix(content,"\n"){content+="\n"};content+=block;return os.WriteFile(path,[]byte(content),0600)}
