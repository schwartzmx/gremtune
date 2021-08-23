package gremcos

// CredentialProvider provides access to cosmos credentials. In order to be able to provide dynamic credentials
// aka cosmos resource tokens you have to implement this interface and ensure in this implementation that always a
// valid resource token is returned by Password().
type CredentialProvider interface {
	Username() (string, error)
	Password() (string, error)
}

// StaticCredentialProvider is a default implementation of the CredentialProvider interface.
// It can be used in case you have no dynamic credentials but use the static primary-/ secondary cosmos key.
type StaticCredentialProvider struct {
	UsernameStatic string
	PasswordStatic string
}

func (c StaticCredentialProvider) Username() (string, error) {
	return c.UsernameStatic, nil
}

func (c StaticCredentialProvider) Password() (string, error) {
	return c.PasswordStatic, nil
}

// noCredentials implementation of the CredentialProvider interface which provides no credentials.
// Is used for unauthenticated connections.
type noCredentials struct {
}

func (c noCredentials) Username() (string, error) {
	return "", nil
}

func (c noCredentials) Password() (string, error) {
	return "", nil
}
