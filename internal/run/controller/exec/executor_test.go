package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasEnvKey(t *testing.T) {
	for name, tt := range map[string]struct {
		env             map[string]string
		caseInsensitive bool
		want            bool
	}{
		"exact key": {
			env:  map[string]string{"CLICOLOR_FORCE": "0"},
			want: true,
		},
		"missing key": {
			env:  map[string]string{"OTHER": "1"},
			want: false,
		},
		"different case is another variable": {
			env:  map[string]string{"clicolor_force": "0"},
			want: false,
		},
		"different case is the same variable on windows": {
			env:             map[string]string{"clicolor_force": "0"},
			caseInsensitive: true,
			want:            true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasEnvKey(tt.env, envClicolorForce, tt.caseInsensitive))
		})
	}
}
