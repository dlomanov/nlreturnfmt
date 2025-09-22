package nlreturnfmt_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dlomanov/nlreturnfmt/pkg/nlreturnfmt"
)

func TestFormatter_FormatBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		blockSize int
		want      string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := os.ReadFile(tt.input)
			require.NoError(t, err)
			want, err := os.ReadFile(tt.want)
			require.NoError(t, err)

			sut := nlreturnfmt.New(nlreturnfmt.WithBlockSize(tt.blockSize))
			got, _, err := sut.FormatFile(tt.input, input)
			require.NoError(t, err)

			assert.Equal(t, string(want), string(got), "formatted output does not match want file for %s", tt.name)
		})
	}
}
