package killer

import (
	"fmt"
	"os"
	"syscall"
)

// Result describes the outcome of a kill operation.
type Result struct {
	Success bool
	PID     int
	Process string
	Signal  string
	Message string
}

// Killer handles terminating target processes safely.
type Killer struct {
	selfPID int
}

// NewKiller creates a Killer instance.
func NewKiller() *Killer {
	return &Killer{
		selfPID: os.Getpid(),
	}
}

// KillProcess attempts to terminate the specified PID with the requested signal.
// If force is true, SIGKILL is sent; otherwise SIGTERM is sent.
func (k *Killer) KillProcess(pid int, procName string, force bool) Result {
	if pid <= 1 {
		return Result{
			Success: false,
			PID:     pid,
			Process: procName,
			Message: fmt.Sprintf("Refusing to terminate critical system PID %d", pid),
		}
	}

	if pid == k.selfPID {
		return Result{
			Success: false,
			PID:     pid,
			Process: procName,
			Message: "Refusing to terminate self (netpulse)",
		}
	}

	sig := syscall.SIGTERM
	sigName := "SIGTERM"
	if force {
		sig = syscall.SIGKILL
		sigName = "SIGKILL"
	}

	err := syscall.Kill(pid, sig)
	if err != nil {
		if os.IsPermission(err) || err == syscall.EPERM {
			return Result{
				Success: false,
				PID:     pid,
				Process: procName,
				Signal:  sigName,
				Message: fmt.Sprintf("Permission denied killing PID %d (run with sudo)", pid),
			}
		}
		if err == syscall.ESRCH {
			return Result{
				Success: true,
				PID:     pid,
				Process: procName,
				Signal:  sigName,
				Message: fmt.Sprintf("Process %d has already exited", pid),
			}
		}
		return Result{
			Success: false,
			PID:     pid,
			Process: procName,
			Signal:  sigName,
			Message: fmt.Sprintf("Failed to kill PID %d: %v", pid, err),
		}
	}

	return Result{
		Success: true,
		PID:     pid,
		Process: procName,
		Signal:  sigName,
		Message: fmt.Sprintf("Killed process %s [%d] with %s", procName, pid, sigName),
	}
}
