package service

import "errors"

var ErrConflict = errors.New("Adding a URL causes a conflict")
