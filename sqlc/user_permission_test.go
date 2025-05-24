package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsInclude(t *testing.T) {
	testCases := []struct {
		name           string
		permissions    []string
		code           string
		expectedResult bool
	}{
		{
			name:           "Permission included",
			permissions:    []string{"read", "write", "delete"},
			code:           "write",
			expectedResult: true,
		},
		{
			name:           "Permission not included",
			permissions:    []string{"read", "write", "delete"},
			code:           "execute",
			expectedResult: false,
		},
		{
			name:           "Empty permissions",
			permissions:    []string{},
			code:           "write",
			expectedResult: false,
		},
		{
			name:           "Nil permissions",
			permissions:    nil,
			code:           "write",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := PermissionsInclude(tc.permissions, tc.code)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
