package constants

import "errors"

// Errors that can occur during bot execution
var (
	ErrStorageKeyNotFound  = errors.New("storage key not found")
	ErrStorageTypeNotFound = errors.New("storage type not found")
	ErrMalformedObject     = errors.New("malformed object type argument")
)

// Errors that are related to a spec
var (
	ErrSpecInvalidNil   = errors.New("invalid spec: nil")
	ErrSpecInvalidType  = errors.New("invalid spec: Type")
	ErrSpecInvalidURI   = errors.New("invalid spec: URI")
)
