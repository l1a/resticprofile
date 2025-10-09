//go:build !darwin && !windows

package schedule

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateSystemdUserJob verifies that when a job is scheduled with "user_logged_on"
// permission, the systemd unit files are created in the user's configuration directory,
// and that the systemctl commands are correctly targeted at the user's instance.
// This is an integration test and requires a running systemd user instance.
func TestCreateSystemdUserJob(t *testing.T) {
	// --- Setup ---
	// Create a job configuration with user-level permission.
	event := calendar.NewEvent()
	err := event.Parse("daily")
	require.NoError(t, err, "Failed to parse calendar event")

	job := &Config{
		ProfileName: "user-profile-integration-test",
		CommandName: "backup",
		Command:     "/usr/local/bin/resticprofile",
		Arguments:   NewCommandArguments([]string{"backup"}),
		// This is the key setting that tells the handler to create a user-level service.
		Permission: constants.SchedulePermissionUserLoggedOn,
		// Add a flag to prevent the job from starting immediately, which is cleaner for testing.
		Flags:     map[string]string{"no-start": ""},
		Schedules: []string{event.String()},
	}
	schedules := []*calendar.Event{event}
	permission := PermissionFromConfig(job.Permission)

	// --- Execution ---
	handler := NewHandler(SchedulerSystemd{}).(*HandlerSystemd)
	err = handler.CreateJob(job, schedules, permission)
	require.NoError(t, err, "CreateJob should not return an error")

	// --- Verification ---
	// Check that the timer and service files were created in the user's home directory.
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	userSystemdPath := filepath.Join(homeDir, ".config", "systemd", "user")
	expectedTimerFile := filepath.Join(userSystemdPath, systemd.GetTimerFile(job.ProfileName, job.CommandName))
	expectedServiceFile := filepath.Join(userSystemdPath, systemd.GetServiceFile(job.ProfileName, job.CommandName))

	assert.FileExists(t, expectedTimerFile, "Timer file should be created in the user's systemd directory")
	assert.FileExists(t, expectedServiceFile, "Service file should be created in the user's systemd directory")

	// --- Cleanup ---
	// Call RemoveJob to clean up the created files and disable the systemd timer.
	err = handler.RemoveJob(job, permission)
	require.NoError(t, err, "RemoveJob should clean up the files without error")

	// Verify that the files have been removed.
	assert.NoFileExists(t, expectedTimerFile, "Timer file should be removed after cleanup")
	assert.NoFileExists(t, expectedServiceFile, "Service file should be removed after cleanup")
}
