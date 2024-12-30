package errors

import "errors"

// HTTP-REST-API Errors
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrPasswordAndConfirmationMismatch = errors.New("password and its confirmation do not match")
var ErrUsernameAlreadyTaken = errors.New("username already taken")
var ErrBadPasswordCriteria = errors.New("a password should have at least 8 chars, 1 uppercase, 1 lowercase, 1 digit and 1 special char")
var ErrBadPUsernameCriteria = errors.New("username must have at least 6 chars")
var ErrInternalServerError = errors.New("ops, an internal server error occurred, this may be unintended btw :( open an issue specifying the steps to reproduce the problem")
var ErrUnableToGetDataFromToken = errors.New("unable to get data from token")
var ErrElementNotFound = errors.New("not found")
var ErrInvalidJSON = errors.New("invalid JSON: ")
var ErrRegistrationNotEnabled = errors.New("registration are not enabled")
var ErrInvalidType = errors.New("failed conversion while fetching db for type: ")
var ErrNoClientFound = errors.New("no client found")
var ErrNotValidClientIP = errors.New("not valid client IP")
var ErrClientIsBusy = errors.New("client is busy")
var ErrCannotDeleteElement = errors.New("unable to delete element")

// GRPC
var ErrGRPCClosedConnection = errors.New("[GRPC]: HashcatChat -> Client has closed the connection ->")
var ErrGRPCFailedToReceive = errors.New("[GRPC]: HashcatChat -> Failed to receive message -> ")
var ErrInvalidToken = errors.New("[GRPC]: HashcatChat -> Invalid token -> ")
var ErrOnUpdateTask = errors.New("[GRPC]: HashcatChat -> Cannot update client task -> ")
var ErrCannotAnswerToClient = errors.New("[GRPC]: HashcatChat -> Cannot reply to the client -> ")
var ErrGetHandshakeStatus = errors.New("[GRPC]: HashcatChat GetHandshakesByStatus -> ")

// SQL
const (
	ErrCodeDuplicateEntry = 1062 // MySQL error code for duplicate entry
)
