package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsVerified(t *testing.T) {
	u := User{
		Active: true,
	}
	assert.True(t, u.IsVerified(), "User should be verified")
}
