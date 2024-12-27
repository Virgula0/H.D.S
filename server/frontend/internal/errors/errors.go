package errors

import "errors"

var ErrLogout = errors.New("unable to logout")
var ErrNotAuthenticated = errors.New("login first")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrResponseFromBackendNotValid = errors.New("not a valid response from backend")
var ErrInvalidJson = errors.New("invalid json in request")
