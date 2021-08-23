package gremcos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticCredentialProvider(t *testing.T) {
	username := "username"
	password := "password"
	provider := StaticCredentialProvider{UsernameStatic: username, PasswordStatic: password}

	uname, err := provider.Username()
	assert.NoError(t, err)
	assert.Equal(t, username, uname)
	pwd, err := provider.Password()
	assert.NoError(t, err)
	assert.Equal(t, password, pwd)
}

func TestNoCredentials(t *testing.T) {
	provider := noCredentials{}

	uname, err := provider.Username()
	assert.NoError(t, err)
	pwd, err := provider.Password()
	assert.NoError(t, err)

	assert.Empty(t, pwd)
	assert.Empty(t, uname)
}
