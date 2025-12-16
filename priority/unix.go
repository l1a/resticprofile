//go:build !windows

package priority

import (
	"errors"
	"fmt"

	"github.com/creativeprojects/clog"
	"golang.org/x/sys/unix"
)

const (
	selfPID      = 0
	errorMessage = "cannot set process group priority to %d, restic will run with the default priority: %w"
)

// SetNice sets the unix "nice" value of the current process
func SetNice(priority int) error {
	var err error

	if priority < -20 || priority > 19 {
		return fmt.Errorf("unexpected priority value %d", priority)
	}

	currentPriority, err := unix.Getpriority(unix.PRIO_PGRP, selfPID)
	if err == nil {
		clog.Debugf("current process group priority is %d", 20-currentPriority)
		if 20-currentPriority >= priority {
			// If the process is already running at a lower priority (higher nice value)
			// than requested, we don't need to change it. This avoids permission errors
			// when running as a normal user who cannot increase priority (lower nice value).
			clog.Debugf("current priority %d is already lower or equal to requested %d, skipping", 20-currentPriority, priority)
			return nil
		}
	}

	clog.Debugf("setting process group priority to %d", priority)
	err = unix.Setpriority(unix.PRIO_PGRP, selfPID, priority)
	if err != nil {
		if errors.Is(err, unix.EPERM) || errors.Is(err, unix.EACCES) {
			// try again after creating a new process group
			return setNewProcessGroup(priority)
		}
		return fmt.Errorf(errorMessage, priority, err)
	}

	return nil
}

func setNewProcessGroup(priority int) error {
	err := unix.Setpgid(selfPID, 0)
	if err != nil {
		return fmt.Errorf("cannot set new process group (priority %d): %w", priority, err)
	}
	err = unix.Setpriority(unix.PRIO_PGRP, selfPID, priority)
	if err != nil {
		return fmt.Errorf("cannot set priority %d on new process group: %w", priority, err)
	}
	return nil
}

// SetClass sets the priority class of the current process
func SetClass(class int) error {
	return SetNice(class)
}
