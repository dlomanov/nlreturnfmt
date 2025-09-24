package main

import (
	"bytes"
	"fmt"
	"log" //nolint: depguard // main_test
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	binaryPath = filepath.Join(os.TempDir(), "nlreturnfmt_test_")
	compile(binaryPath)

	exitCode := m.Run()

	_ = os.Remove(binaryPath)
	os.Exit(exitCode)
}

func TestCLI(t *testing.T) {
	input := readFile(t, "../../testdata/p/p.input.go")
	golden := readFile(t, "../../testdata/p/p.golden.go")

	tests := []struct {
		name         string
		args         []string
		stdin        string
		setup        func(t *testing.T) (filePath string, teardown func())
		wantExitCode int
		wantStdout   string
		wantStderr   string
	}{
		{
			name:         "-version flag",
			args:         []string{"-version"},
			wantExitCode: 0,
			wantStdout:   "nlreturnfmt version dev",
		},
		{
			name: "format file to stdout",
			setup: func(t *testing.T) (string, func()) {
				return writeFile(t, "test.go", input), nil
			},
			args:         []string{},
			wantExitCode: 0,
			wantStdout:   fmt.Sprintf("// %s - formatted:\n%s\n", "<filepath>", golden),
		},
		{
			name: "format file with -w flag",
			setup: func(t *testing.T) (string, func()) {
				filePath := writeFile(t, "test.go", input)

				return filePath, func() {
					got := readFile(t, filePath)
					require.Equal(t, string(golden), string(got))
				}
			},
			args:         []string{"-w"},
			wantExitCode: 0,
			wantStdout:   "",
		},
		{
			name:         "error on non-existent file",
			args:         []string{"non_existent_file.go"},
			wantExitCode: 1,
			wantStderr:   "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.setup != nil {
				var teardown func()
				filePath, teardown = tt.setup(t)
				if teardown != nil {
					defer teardown()
				}
			}

			args := tt.args
			if filePath != "" {
				args = append(args, filePath)
			}

			cmd := exec.Command(binaryPath, args...)
			cmd.Stdin = strings.NewReader(tt.stdin)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.wantExitCode == 0 {
				require.NoError(t, err, "command failed unexpectedly. Stderr: %s", stderr.String())
			} else {
				require.Error(t, err, "command should have failed but didn't")
				var exitErr *exec.ExitError
				require.ErrorAs(t, err, &exitErr, "error is not an exec.ExitError")
				require.Equal(t, tt.wantExitCode, exitErr.ExitCode(), "unexpected exit code")
			}

			wantStdout := tt.wantStdout
			if strings.Contains(wantStdout, "<filepath>") {
				wantStdout = strings.ReplaceAll(wantStdout, "<filepath>", filePath)
			}
			require.Contains(t, stdout.String(), wantStdout)
			require.Contains(t, stderr.String(), tt.wantStderr)
		})
	}
}

func readFile(t *testing.T, path string) []byte {
	v, err := os.ReadFile(path)
	require.NoError(t, err)

	return v
}

func writeFile(t *testing.T, path string, content []byte) string {
	filePath := filepath.Join(t.TempDir(), path)
	err := os.WriteFile(filePath, content, 0o644)
	require.NoError(t, err)

	return filePath
}

func compile(path string) {
	buildCmd := exec.Command("go", "build", "-o", path, ".")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		log.Fatalf("buildCmd.CombinedOutput(): %s\n%v", string(output), err)
	}
}
