// Package decoder provides decoding utilities.
package decoder

import "io"

// Decoder decodes contents of a [io.Reader] into a variable.
type Decoder func(io.Reader, any) error
