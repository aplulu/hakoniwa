package model

import "errors"

var (
	ErrUnauthorized        = errors.New("unauthorized")
	ErrNotFound            = errors.New("not found")
	ErrMaxInstancesReached = errors.New("max instances reached")
)
