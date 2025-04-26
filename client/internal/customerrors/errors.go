package customerrors

import (
	"database/sql"
	"errors"
)

var ErrFinalSending = errors.New("[CLIENT] Failed to send final status, retrying in ")
var ErrHcxToolsNotFound = errors.New("conversion was not successful, hcxtools output file not found")
var ErrNoRowsFound = sql.ErrNoRows

var ErrInternalServerError = errors.New("internal server error")
