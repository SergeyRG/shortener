package service

import "errors"

var ErrConflict = errors.New("adding a URL causes a conflict")
