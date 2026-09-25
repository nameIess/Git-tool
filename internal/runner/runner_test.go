package runner

import "testing"

func TestResultSuccess(t *testing.T) {
	if !(Result{ExitCode: 0}).Success() { t.Fatal("zero exit code without error should succeed") }
	if (Result{ExitCode: 1}).Success() { t.Fatal("non-zero exit code should fail") }
}

func TestCombinedOutput(t *testing.T) {
	got := (Result{Stdout: " out ", Stderr: " err "}).CombinedOutput()
	if got != "out\nerr" { t.Fatalf("unexpected combined output: %q", got) }
}
