package system

import "runtime"

const (
	// https://serverfault.com/questions/69430/what-is-the-maximum-length-of-a-command-line-in-mac-os-x
	// https://support.microsoft.com/en-us/help/830473/command-prompt-cmd-exe-command-line-string-limitation
	// https://unix.stackexchange.com/a/120652
	maxCommandLengthDarwin  = 260000 // 262144
	maxCommandLengthWindows = 7000   // 8191, but see issues#655
	maxCommandLengthLinux   = 130000 // 131072
)

func MaxCmdLen() int {
	switch runtime.GOOS {
	case "windows":
		return maxCommandLengthWindows
	case "darwin":
		return maxCommandLengthDarwin
	default:
		return maxCommandLengthLinux
	}
}
