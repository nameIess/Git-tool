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

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

func (r Result) Success() bool {
	return r.Err == nil && r.ExitCode == 0
}

func (r Result) CombinedOutput() string {
	parts := []string{}
	if r.Stdout != "" {
		parts = append(parts, strings.TrimSpace(r.Stdout))
	}
	if r.Stderr != "" {
		parts = append(parts, strings.TrimSpace(r.Stderr))
	}
	return strings.Join(parts, "\n")
}

func Run(name string, args ...string) Result {
	return RunCtx(context.Background(), name, args...)
}

func RunCtx(ctx context.Context, name string, args ...string) Result {
	logger.Debug("Executing: %s %s", name, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if err != nil {
		logger.Debug("Command failed (exit %d): %v\nStderr: %s", result.ExitCode, err, result.Stderr)
	} else {
		logger.Debug("Command succeeded. Stdout: %s", strings.TrimSpace(result.Stdout))
	}

	return result
}

func RunWithStdin(input string, name string, args ...string) Result {
	logger.Debug("Executing (with stdin): %s %s", name, strings.Join(args, " "))

	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if err != nil {
		logger.Debug("Command failed (exit %d): %v", result.ExitCode, err)
	} else {
		logger.Debug("Command succeeded.")
	}

	return result
}

func RunWithTimeout(timeout time.Duration, name string, args ...string) Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunCtx(ctx, name, args...)
}

func Which(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	return path, nil
}
