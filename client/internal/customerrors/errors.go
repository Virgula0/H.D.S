package customerrors

import "errors"

var ErrFinalSending = errors.New("[CLIENT] Failed to send final status, retrying in ")
var ErrHcxToolsNotFound = errors.New("conversion was not successful, hcxtools output file not found")
