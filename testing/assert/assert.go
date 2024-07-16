// Package assert provides basic assertion utilities.
package assert

import (
	"reflect"
	"testing"
)

// Note: this file contains some assertion utilities.

// Equal checks if 2 values are equal.
// Returns successful assertion status.
func Equal(t *testing.T, expected any, actual any) bool {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(
			"\n\t"+`expected "%+v" (%T),`+
				"\n\t"+`but got  "%+v" (%T)`+"\n",
			expected, expected,
			actual, actual,
		)

		return false
	}

	return true
}

// NotNil checks if value passed is not nil.
// Returns successful assertion status.
func NotNil(t *testing.T, actual any) bool {
	t.Helper()
	if isNil(actual) {
		t.Error("should not be nil")

		return false
	}

	return true
}

// Nil checks if value passed is nil.
// Returns successful assertion status.
func Nil(t *testing.T, actual any) bool {
	t.Helper()
	if !isNil(actual) {
		t.Errorf("expected nil, but got %+v", actual)

		return false
	}

	return true
}

// RequireNil fails the test immediately if passed value is not nil.
func RequireNil(t *testing.T, actual any) {
	t.Helper()
	if !isNil(actual) {
		t.Errorf("expected nil, but got %+v", actual)
		t.FailNow()
	}
}

// True checks if value passed is true.
// Returns successful assertion status.
func True(t *testing.T, actual bool) bool {
	t.Helper()
	if !actual {
		t.Error("should be true")

		return false
	}

	return true
}

// isNil checks an interface if it is nil.
func isNil(object any) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)

	kind := value.Kind()
	switch kind {
	case reflect.Ptr:
		return value.IsNil()
	case reflect.Slice:
		return value.IsNil()
	case reflect.Map:
		return value.IsNil()
	case reflect.Interface:
		return value.IsNil()
	case reflect.Func:
		return value.IsNil()
	case reflect.Chan:
		return value.IsNil()
	}

	return false
}
