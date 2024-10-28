package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input     string
		want      []string
		shouldErr bool
	}{
		{
			input: "-itd",
			want:  []string{"-i", "-t", "-d"},
		},
		{
			input: "-itf=b",
			want:  []string{"-i", "-t", "-f=b"},
		},
		{
			input:     "-itd#",
			shouldErr: true,
		},
		{
			input: "-f=b",
			want:  []string{"-f=b"},
		},
		{
			input:     "-=b",
			shouldErr: true,
		},
		{
			input: "-",
			want:  []string{"-"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			got, err := parseArgs([]string{"rancher", "run", "--debug", test.input})
			if test.shouldErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, append([]string{"rancher", "run", "--debug"}, test.want...), got)
		})
	}
}
