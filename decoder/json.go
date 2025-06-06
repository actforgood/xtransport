package decoder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// DecodeJSON decodes JSON contents of a [io.Reader] into a variable.
func DecodeJSON(r io.Reader, dest any) error {
	dec := json.NewDecoder(r)
	// dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		var (
			syntaxError        *json.SyntaxError
			unmarshalTypeError *json.UnmarshalTypeError
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("badly-formed JSON (at position %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf(
				"invalid value for the %q field (at position %d)",
				unmarshalTypeError.Field,
				unmarshalTypeError.Offset,
			)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")

			return fmt.Errorf("unknown field %s", fieldName)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case err.Error() == "http: request body too large":
			return errors.New("body too large")
		default:
			return err
		}
	}
	err := dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("json does not contain a single object")
	}

	return nil
}
