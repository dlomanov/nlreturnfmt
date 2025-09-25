package nlreturnfmt_test

import (
	"os"
	"testing"

	"github.com/dlomanov/nlreturnfmt/pkg/nlreturnfmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_FormatBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		blockSize int
		want      string
		wantErr   bool
	}{
		{
			name:      "p",
			blockSize: 1,
			input:     "../../testdata/p/p.input.go",
			want:      "../../testdata/p/p.golden.go",
		},
		{
			name:      "bs",
			blockSize: 2,
			input:     "../../testdata/bs/bs.input.go",
			want:      "../../testdata/bs/bs.golden.go",
		},
		{
			name:      "idempotent",
			blockSize: 2,
			input:     "../../testdata/idempotent/idempotent.go",
			want:      "../../testdata/idempotent/idempotent.go",
		},
		{
			name:      "branches",
			blockSize: 1,
			input:     "../../testdata/branches/branches.input.go",
			want:      "../../testdata/branches/branches.golden.go",
		},
		{
			name:      "closures",
			blockSize: 1,
			input:     "../../testdata/closures/closures.input.go",
			want:      "../../testdata/closures/closures.golden.go",
		},
		{
			// This test verifies that a statement preceded by a comment is not modified.
			// This is intentional, to align with the original nlreturn linter's behavior.
			name:      "comments",
			blockSize: 1,
			input:     "../../testdata/comments/comments.go",
			want:      "../../testdata/comments/comments.go",
		},
		{
			name:      "syntax error",
			blockSize: 1,
			input:     "../../testdata/errors/syntax_error.input.go",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := read(t, tt.input)
			want := read(t, tt.want)

			sut := nlreturnfmt.New(nlreturnfmt.WithBlockSize(tt.blockSize))
			got, _, err := sut.FormatFile(t.Context(), tt.input, input)

			if tt.wantErr {
				require.Error(t, err, "expected an error for test case: %s", tt.name)
			} else {
				require.NoError(t, err)
				assert.Equal(t, string(want), string(got), "formatted output does not match want file for %s", tt.name)
			}
		})
	}
}

func read(t *testing.T, filename string) []byte {
	if filename == "" {
		return nil
	}

	content, err := os.ReadFile(filename)
	require.NoError(t, err)

	return content
}
