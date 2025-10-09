package envscanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanForEnvVariables(t *testing.T) {
	yamlContent := `
global:
  restic-binary: /usr/local/bin/restic

profiles:
  profile-with-vars:
    repository: /data/.Env.REPO_SUFFIX
    password-file: .Env.HOME/.restic-password
    source:
      - .Env.HOME/documents
      - /srv/data
    env:
      AWS_KEY: .Env.MY_AWS_KEY

  profile-no-vars:
    repository: /data/no-vars-repo
    source:
      - /srv/other_data

groups:
  group-with-vars:
    profiles:
      - profile-with-vars
    schedules:
      backup:
        at: "*-*-* 01:00:00"
        env:
          EXTRA_VAR: .Env.GROUP_VAR
`
	// Create a temporary config file for the test
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "profiles.yaml")
	err := os.WriteFile(configFile, []byte(yamlContent), 0600)
	require.NoError(t, err, "Failed to write temporary config file")

	testCases := []struct {
		name               string
		profileOrGroupName string
		expectedVars       []string
		expectError        bool
	}{
		{
			name:               "Profile with multiple and duplicate variables",
			profileOrGroupName: "profile-with-vars",
			expectedVars:       []string{"HOME", "MY_AWS_KEY", "REPO_SUFFIX"},
			expectError:        false,
		},
		{
			name:               "Profile with no variables",
			profileOrGroupName: "profile-no-vars",
			expectedVars:       nil, // Expect nil or empty slice
			expectError:        false,
		},
		{
			name:               "Group with its own variables",
			profileOrGroupName: "group-with-vars",
			expectedVars:       []string{"GROUP_VAR"},
			expectError:        false,
		},
		{
			name:               "Profile that does not exist",
			profileOrGroupName: "non-existent-profile",
			expectedVars:       nil,
			expectError:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vars, err := ScanForEnvVariables(configFile, tc.profileOrGroupName)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if len(tc.expectedVars) == 0 {
					assert.Empty(t, vars)
				} else {
					assert.Equal(t, tc.expectedVars, vars)
				}
			}
		})
	}
}

func TestScanForEnvVariables_FileNotFound(t *testing.T) {
	_, err := ScanForEnvVariables("/path/to/non/existent/file.yaml", "any-profile")
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to read config file")
}

func TestScanForEnvVariables_InvalidYaml(t *testing.T) {
	// Create a temporary invalid config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(configFile, []byte("profiles: \n  profile-1: [invalid"), 0600)
	require.NoError(t, err)

	_, err = ScanForEnvVariables(configFile, "profile-1")
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to unmarshal YAML")
}
