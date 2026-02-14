package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/user/git-tool/internal/logger"
)

// Result holds the output of a command execution.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Success returns true if the command completed without error.
func (r Result) Success() bool {
	return r.Err == nil && r.ExitCode == 0
}

// CombinedOutput returns stdout + stderr concatenated.
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

// Run executes a command and returns the result.
func Run(name string, args ...string) Result {
	return RunCtx(context.Background(), name, args...)
}

// RunCtx executes a command with context.
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
		logger.Debug("Command failed (exit %d): %v", result.ExitCode, err)
		logger.Debug("Stderr: %s", result.Stderr)
	} else {
		logger.Debug("Command succeeded. Stdout: %s", strings.TrimSpace(result.Stdout))
	}

	return result
}

// RunWithStdin executes a command, writing input to stdin.
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

// RunWithTimeout executes a command with a timeout.
func RunWithTimeout(timeout time.Duration, name string, args ...string) Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunCtx(ctx, name, args...)
}

// Which checks if a command is available in PATH and returns its path.
func Which(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH: %w", name, err)
	}
	return path, nil
}
