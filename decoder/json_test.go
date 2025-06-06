package decoder_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/actforgood/xtransport/decoder"
	"github.com/actforgood/xtransport/testing/assert"
)

type testDummyStruct struct {
	Name string
}

func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	t.Run("successfully decodes json", testDecodeJSONSuccess)
	t.Run("returns misc error", testDecodeJSONErrMisc)
	t.Run("returns error", testDecodeJSONErr)
}

func testDecodeJSONSuccess(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		r            = strings.NewReader(`[1,2,3]`)
		dest         []byte
		expectedDest = []byte{1, 2, 3}
	)

	// act
	err := decoder.DecodeJSON(r, &dest)

	// assert
	if assert.Nil(t, err) {
		assert.Equal(t, expectedDest, dest)
	}
}

func testDecodeJSONErrMisc(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		r    = strings.NewReader(`{"Name":"John Doe"}`)
		dest *testDummyStruct
	)

	// act
	err := decoder.DecodeJSON(r, dest)

	// assert
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), `json: Unmarshal`))
	}
}

func testDecodeJSONErr(t *testing.T) {
	t.Parallel()

	// arrange
	var dest testDummyStruct
	tests := [...]struct {
		name          string
		reader        io.Reader
		expectedError string
	}{
		{
			name:          "returns empty body error",
			reader:        strings.NewReader(""),
			expectedError: "body must not be empty",
		},
		{
			name:          "returns multi objects error",
			reader:        strings.NewReader(`{"Name":"John Doe"}{"Name":"Jane Doe"}`),
			expectedError: "json does not contain a single object",
		},
		{
			name:          "returns invalid field type error",
			reader:        strings.NewReader(`{"Name":123}`),
			expectedError: `invalid value for the "Name" field (at position 11)`,
		},
		// {
		// 	name:          "returns extra field error",
		// 	reader:        strings.NewReader(`{"Name":"John Doe","ExtraField":"trigger error"}`),
		// 	expectedError: `unknown field "ExtraField"`,
		// },
		{
			name:          "returns malformed json error at position",
			reader:        strings.NewReader(`{{"Name":"John Doe"}`),
			expectedError: "badly-formed JSON (at position 2)",
		},
		{
			name:          "returns malformed json error generic",
			reader:        strings.NewReader(`{"Name":"John Doe"`),
			expectedError: "badly-formed JSON",
		},
		{
			name:          "returns body too large error",
			reader:        http.MaxBytesReader(nil, io.NopCloser(strings.NewReader(`{"Name":"John Doe"}`)), 1),
			expectedError: "body too large",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// act
			actualError := decoder.DecodeJSON(test.reader, &dest)

			// assert
			if assert.NotNil(t, actualError) {
				assert.Equal(t, test.expectedError, actualError.Error())
			}
		})
	}
}
