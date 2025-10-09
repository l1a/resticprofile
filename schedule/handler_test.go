package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestLookupExistingBinary(t *testing.T) {
	if platform.IsWindows() {
		t.Skip()
	}
	err := lookupBinary("sh", "sh")
	assert.NoError(t, err)
}

func TestLookupNonExistingBinary(t *testing.T) {
	if platform.IsWindows() {
		t.Skip()
	}
	err := lookupBinary("something", "almost_certain_not_to_be_available")
	assert.Error(t, err)
}

func TestPermissions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		permission Permission
		euid       int
		expected   bool
	}{
		{
			name:       "PermissionUserLoggedOn",
			permission: PermissionUserLoggedOn,
			euid:       0,
			expected:   true,
		},
		{
			name:       "PermissionUserLoggedOn",
			permission: PermissionUserLoggedOn,
			euid:       1,
			expected:   true,
		},
		{
			name:       "PermissionUserBackground",
			permission: PermissionUserBackground,
			euid:       0,
			expected:   true,
		},
		{
			name:       "PermissionUserBackground",
			permission: PermissionUserBackground,
			euid:       1,
			expected:   true,
		},
		{
			name:       "PermissionSystem",
			permission: PermissionSystem,
			euid:       0,
			expected:   true,
		},
		{
			name:       "PermissionSystem",
			permission: PermissionSystem,
			euid:       1,
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CheckPermission(tc.permission, tc.euid))
		})
	}
}
