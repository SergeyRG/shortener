package errors

import "errors"

var (
	ErrAlredyExist = errors.New("short URL with this ID already exist")
	ErrURLNotFound = errors.New("URL not found by ID")
	ErrUnexpected  = errors.New("unexpected error")
	ErrNotEnoughID = errors.New("cant create a short link. Not enough available IDs")
)
