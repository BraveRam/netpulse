package killer

import (
	"os"
	"testing"
)

func TestKillerGuards(t *testing.T) {
	k := NewKiller()

	// Should reject PID 0
	res0 := k.KillProcess(0, "init", false)
	if res0.Success {
		t.Errorf("Expected failure for PID 0")
	}

	// Should reject PID 1
	res1 := k.KillProcess(1, "systemd", false)
	if res1.Success {
		t.Errorf("Expected failure for PID 1")
	}

	// Should reject self PID
	selfPID := os.Getpid()
	resSelf := k.KillProcess(selfPID, "netpulse", false)
	if resSelf.Success {
		t.Errorf("Expected failure for self PID")
	}
}
