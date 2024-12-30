package errors

import "errors"

var ErrLogout = errors.New("unable to logout")
var ErrNotAuthenticated = errors.New("login first")
var ErrInvalidCredentials = errors.New("invalid credentials")
