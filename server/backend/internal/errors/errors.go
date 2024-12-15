package errors

import "errors"

// HTTP-REST-API Errors
var ErrInvalidCreds = errors.New("invalid credentials")
var ErrPaswwordAndConfirmationMismatch = errors.New("password and its confirmation do not match")
var ErrUsernameAlreadyTaken = errors.New("username already taken")
var ErrBadPasswordCriteria = errors.New("a password should have at least 8 chars, 1 uppercase, 1 lowercase, 1 digit and 1 special char")
var ErrBadPUsernameCriteria = errors.New("username must have at least 6 chars")
var ErrInternalServerError = errors.New("ops, an internal server error occurred, this may be unintented btw :( report the error on discord")
var ErrUnableToGetDataFromToken = errors.New("unable to get data from token")
var ErrElementNotFound = errors.New("not found")
var ErrInvalidJSON = errors.New("invalid JSON: ")
var ErrRegistrationNotEnabled = errors.New("registration are not enabled")
var ErrInvalidType = errors.New("failed convertsion while fetching db for type ")
