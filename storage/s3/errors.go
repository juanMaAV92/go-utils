package s3

import "errors"

// ErrObjectNotFound is returned by HeadObject when the key does not exist.
var ErrObjectNotFound = errors.New("s3: object not found")
