package gremcos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticCredentialProvider(t *testing.T) {
	username := "username"
	password := "password"
	provider := StaticCredentialProvider{UsernameStatic: username, PasswordStatic: password}

	assert.Equal(t, username, provider.Username())
	assert.Equal(t, password, provider.Password())
}

func TestNoCredentials(t *testing.T) {
	provider := noCredentials{}

	assert.Empty(t, provider.Username())
	assert.Empty(t, provider.Password())
}
