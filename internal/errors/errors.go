package errors

import "errors"

var (
	AlredyExistError = errors.New("Short URL with this ID already exist")
	URLNotFoundError = errors.New("URL not found by ID")
)
