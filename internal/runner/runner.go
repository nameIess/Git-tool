package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nameIess/git-tool/internal/logger"
)

type Result struct { Stdout, Stderr string; ExitCode int; Err error }
func (r Result) Success() bool{return r.Err==nil&&r.ExitCode==0}
func (r Result) CombinedOutput()string{var p []string;if strings.TrimSpace(r.Stdout)!=""{p=append(p,strings.TrimSpace(r.Stdout))};if strings.TrimSpace(r.Stderr)!=""{p=append(p,strings.TrimSpace(r.Stderr))};return strings.Join(p,"\n")}
func Run(name string,args ...string)Result{return RunCtx(context.Background(),name,args...)}
func RunRedacted(name string,redacted []int,args ...string)Result{return run(context.Background(),name,redacted,args...)}
func RunCtx(ctx context.Context,name string,args ...string)Result{return run(ctx,name,nil,args...)}
func run(ctx context.Context,name string,redacted []int,args ...string)Result{display:=append([]string(nil),args...);for _,i:=range redacted{if i>=0&&i<len(display){display[i]="<redacted>"}};logger.Debug("Executing: %s %s",name,strings.Join(display," "));cmd:=exec.CommandContext(ctx,name,args...);var out,errout bytes.Buffer;cmd.Stdout=&out;cmd.Stderr=&errout;err:=cmd.Run();r:=Result{Stdout:out.String(),Stderr:errout.String(),Err:err};if cmd.ProcessState!=nil{r.ExitCode=cmd.ProcessState.ExitCode()};if err!=nil{logger.Debug("Command failed (exit %d): %v",r.ExitCode,err)};return r}
func RunWithStdin(input,name string,args ...string)Result{cmd:=exec.Command(name,args...);cmd.Stdin=strings.NewReader(input);var out,errout bytes.Buffer;cmd.Stdout=&out;cmd.Stderr=&errout;err:=cmd.Run();r:=Result{Stdout:out.String(),Stderr:errout.String(),Err:err};if cmd.ProcessState!=nil{r.ExitCode=cmd.ProcessState.ExitCode()};return r}
func RunWithTimeout(timeout time.Duration,name string,args ...string)Result{ctx,cancel:=context.WithTimeout(context.Background(),timeout);defer cancel();return RunCtx(ctx,name,args...)}
func Which(name string)(string,error){p,err:=exec.LookPath(name);if err!=nil{return "",fmt.Errorf("%s not found in PATH: %w",name,err)};return p,nil}
