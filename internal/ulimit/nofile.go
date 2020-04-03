package ulimit

import (
	"fmt"
	"runtime"
	"syscall"
)

func getNoFileLimit() (syscall.Rlimit, error) {
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return rLimit, fmt.Errorf("getrlimit: %w", err)
	}
	return rLimit, nil
}

// SetNoFileLimitToMax changes the open file descriptor limit of the current
// process to the available maximum.
func SetNoFileLimitToMax() (before, current, max uint64, err error) {
	rLimitBefore, err := getNoFileLimit()
	if err != nil {
		return 0, 0, 0, err
	}

	if rLimitBefore.Cur == rLimitBefore.Max {
		// max is already configured, do nothing
		return rLimitBefore.Cur, rLimitBefore.Cur, rLimitBefore.Max, nil
	}

	// BUG(tisba): There seems to be an issue starting with Go 1.12
	// See https://github.com/golang/go/issues/30401#issuecomment-467530109
	rLimit := rLimitBefore
	if runtime.GOOS == "darwin" {
		if rLimit.Cur > 10240 {
			// The max file limit is 10240, even though
			// the max returned by Getrlimit is 1<<63-1.
			// This is OPEN_MAX in sys/syslimits.h.
			rLimit.Cur = 10240
		}
	} else {
		rLimit.Cur = rLimit.Max
	}

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("setrlimit: %w", err)
	}

	rLimit, err = getNoFileLimit()
	if err != nil {
		return 0, 0, 0, err
	}

	return rLimitBefore.Cur, rLimit.Cur, rLimit.Max, nil
}
