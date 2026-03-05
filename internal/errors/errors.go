package errors

import "errors"

var (
	ErrAlredyExist = errors.New("short URL with this ID already exist")
	ErrURLNotFound = errors.New("URL not found by ID")
)
