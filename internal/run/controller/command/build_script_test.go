package command

import "testing"

func TestQuoteScriptPath(t *testing.T) {
	t.Parallel()

	path := `C:\somepath\with spaces\.lefthook\commit-msg\validate_format.ps1`

	tests := []struct {
		name string
		goos string
		want string
	}{
		{
			name: "unix",
			goos: "linux",
			want: `'C:\somepath\with spaces\.lefthook\commit-msg\validate_format.ps1'`,
		},
		{
			name: "windows",
			goos: "windows",
			want: `'"C:\somepath\with spaces\.lefthook\commit-msg\validate_format.ps1"'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := quoteScriptPath(path, tt.goos); got != tt.want {
				t.Errorf("quoteScriptPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
