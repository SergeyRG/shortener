package repository

import "errors"

var (
	ErrAlreadyExist = errors.New("short URL with this ID already exist")
	ErrURLNotFound  = errors.New("URL not found by ID")
	ErrUnexpected   = errors.New("unexpected error")
	ErrNotEnoughID  = errors.New("cant create a short link. Not enough available IDs")
)
