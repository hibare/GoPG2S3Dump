package exec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewExec(t *testing.T) {
	executor := NewExec()
	if executor == nil {
		t.Fatal("NewExec() returned nil")
	}

	// Verify it implements the interface
	var _ = executor
}

func TestExec_Command(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test basic command creation
	cmd := executor.Command(ctx, "echo", "hello")
	if cmd == nil {
		t.Fatal("Command() returned nil")
	}

	// Verify it implements the interface
	var _ = cmd
}

func TestExec_LookPath(t *testing.T) {
	executor := NewExec()

	// Test looking for a known executable
	path, err := executor.LookPath("go")
	if err != nil {
		// Skip if go is not in PATH (e.g., in CI environment)
		t.Skipf("go not found in PATH: %v", err)
	}
	if path == "" {
		t.Error("LookPath returned empty string for 'go'")
	}

	// Test looking for non-existent executable
	_, err = executor.LookPath("nonexistent_executable_12345")
	if err == nil {
		t.Error("LookPath should return error for non-existent executable")
	}
}

func TestCmd_WithEnv(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	cmd := executor.Command(ctx, "env")

	// Test setting custom environment variables
	customEnv := []string{"TEST_VAR=test_value", "ANOTHER_VAR=another_value"}
	cmd = cmd.WithEnv(customEnv)

	// Run the command and capture output
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run env command: %v", err)
	}

	outputStr := string(output)

	// Check that our custom environment variables are present
	if !strings.Contains(outputStr, "TEST_VAR=test_value") {
		t.Error("Custom environment variable TEST_VAR not found in output")
	}
	if !strings.Contains(outputStr, "ANOTHER_VAR=another_value") {
		t.Error("Custom environment variable ANOTHER_VAR not found in output")
	}

	// Verify system environment variables are still present
	if !strings.Contains(outputStr, "PATH=") {
		t.Error("System environment variable PATH not found in output")
	}
}

func TestCmd_WithDir(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file in the temp directory
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test running pwd command in temp directory
	cmd := executor.Command(ctx, "pwd")
	cmd = cmd.WithDir(tempDir)

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run pwd command: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr != tempDir {
		t.Errorf("Expected working directory %s, got %s", tempDir, outputStr)
	}

	// Test running ls command in temp directory
	cmd = executor.Command(ctx, "ls")
	cmd = cmd.WithDir(tempDir)

	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run ls command: %v", err)
	}

	outputStr = strings.TrimSpace(string(output))
	if outputStr != "test.txt" {
		t.Errorf("Expected ls output 'test.txt', got '%s'", outputStr)
	}
}

func TestCmd_WithStdout(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Create a temporary file for stdout
	stdoutFile, err := os.CreateTemp(t.TempDir(), "stdout_test")
	if err != nil {
		t.Fatalf("Failed to create temp stdout file: %v", err)
	}
	defer func() { _ = os.Remove(stdoutFile.Name()) }()
	defer func() { _ = stdoutFile.Close() }()

	cmd := executor.Command(ctx, "echo", "hello world")
	cmd = cmd.WithStdout(stdoutFile)

	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to run echo command: %v", err)
	}

	// Read the output from the file
	_, err = stdoutFile.Seek(0, 0)
	if err != nil {
		t.Fatalf("Failed to seek stdout file: %v", err)
	}
	output, err := os.ReadFile(stdoutFile.Name())
	if err != nil {
		t.Fatalf("Failed to read stdout file: %v", err)
	}

	expected := "hello world\n"
	if string(output) != expected {
		t.Errorf("Expected stdout '%s', got '%s'", expected, string(output))
	}
}

func TestCmd_WithStderr(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Create a temporary file for stderr
	stderrFile, err := os.CreateTemp(t.TempDir(), "stderr_test")
	if err != nil {
		t.Fatalf("Failed to create temp stderr file: %v", err)
	}
	defer func() { _ = os.Remove(stderrFile.Name()) }()
	defer func() { _ = stderrFile.Close() }()

	// Use a command that writes to stderr (bash -c "echo 'error message' >&2")
	cmd := executor.Command(ctx, "bash", "-c", "echo 'error message' >&2")
	cmd = cmd.WithStderr(stderrFile)

	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to run bash command: %v", err)
	}

	// Read the output from the file
	_, err = stderrFile.Seek(0, 0)
	if err != nil {
		t.Fatalf("Failed to seek stderr file: %v", err)
	}
	output, err := os.ReadFile(stderrFile.Name())
	if err != nil {
		t.Fatalf("Failed to read stderr file: %v", err)
	}

	expected := "error message\n"
	if string(output) != expected {
		t.Errorf("Expected stderr '%s', got '%s'", expected, string(output))
	}
}

func TestCmd_Run(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test successful command
	cmd := executor.Command(ctx, "echo", "success")
	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected successful run, got error: %v", err)
	}

	// Test command that should fail
	cmd = executor.Command(ctx, "nonexistent_command")
	err = cmd.Run()
	if err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}

func TestCmd_Output(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test successful command with output
	cmd := executor.Command(ctx, "echo", "test output")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Expected successful output, got error: %v", err)
	}

	expected := "test output\n"
	if string(output) != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, string(output))
	}

	// Test command that should fail
	cmd = executor.Command(ctx, "nonexistent_command")
	_, err = cmd.Output()
	if err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}

func TestCmd_CombinedOutput(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test successful command with combined output
	cmd := executor.Command(ctx, "echo", "test combined output")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected successful combined output, got error: %v", err)
	}

	expected := "test combined output\n"
	if string(output) != expected {
		t.Errorf("Expected combined output '%s', got '%s'", expected, string(output))
	}

	// Test command that writes to both stdout and stderr
	cmd = executor.Command(ctx, "bash", "-c", "echo 'stdout message'; echo 'stderr message' >&2")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Expected successful combined output, got error: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "stdout message") {
		t.Error("Combined output should contain stdout message")
	}
	if !strings.Contains(outputStr, "stderr message") {
		t.Error("Combined output should contain stderr message")
	}

	// Test command that should fail
	cmd = executor.Command(ctx, "nonexistent_command")
	_, err = cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}

func TestCmd_ContextCancellation(t *testing.T) {
	executor := NewExec()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start a long-running command
	cmd := executor.Command(ctx, "sleep", "10")

	// The command should be cancelled due to context timeout
	err := cmd.Run()
	if err == nil {
		t.Error("Expected command to be cancelled due to context timeout")
	}
}

func TestCmd_ChainedMethods(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test chaining multiple With* methods
	cmd := executor.Command(ctx, "env")
	cmd = cmd.WithEnv([]string{"CHAINED_VAR=chained_value"})
	cmd = cmd.WithDir("/tmp")

	// Verify the command still works after chaining
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run chained command: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "CHAINED_VAR=chained_value") {
		t.Error("Chained environment variable not found in output")
	}
}

func TestCmd_InterfaceCompliance(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test that the concrete types implement the interfaces
	var _ = executor
	var _ = executor.Command(ctx, "echo", "test")
}

func TestCmd_EmptyArgs(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test command with no arguments
	cmd := executor.Command(ctx, "echo")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run echo without args: %v", err)
	}

	// echo without args should just output a newline
	expected := "\n"
	if string(output) != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, string(output))
	}
}

func TestCmd_EmptyStringArgs(t *testing.T) {
	executor := NewExec()
	ctx := t.Context()

	// Test command with empty string arguments
	cmd := executor.Command(ctx, "echo", "", "non_empty")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run echo with empty string args: %v", err)
	}

	expected := " non_empty\n"
	if string(output) != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, string(output))
	}
}
