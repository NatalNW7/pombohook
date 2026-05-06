package cli

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelperProcess is a dummy process that just sleeps.
// It is used to test process management without relying on external commands.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Sleep for 10 seconds to keep the process alive
	time.Sleep(10 * time.Second)
	os.Exit(0)
}

func TestProcessManagement(t *testing.T) {
	// Start a dummy process
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	cmd.SysProcAttr = sysProcAttr()

	err := cmd.Start()
	require.NoError(t, err)

	pid := cmd.Process.Pid

	// Prevent the process from becoming a zombie
	go func() {
		_ = cmd.Wait()
	}()

	// Ensure the process is marked as alive
	assert.True(t, isProcessAlive(pid), "Process should be alive after starting")

	// Stop the process
	err = stopProcess(pid)
	require.NoError(t, err)

	// Allow a small window for the OS to clean up the process state
	time.Sleep(200 * time.Millisecond)

	// Ensure the process is marked as dead
	assert.False(t, isProcessAlive(pid), "Process should be dead after stopping")
}

func TestStopNonExistentProcess(t *testing.T) {
	// Use an extremely high PID that is highly unlikely to exist
	pid := 999999

	err := stopProcess(pid)
	assert.Error(t, err, "Should return an error when stopping a non-existent process")

	assert.False(t, isProcessAlive(pid), "Non-existent process should not be alive")
}
