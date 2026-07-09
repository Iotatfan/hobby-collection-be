package common

import (
	"errors"
	"net/http"
	"testing"

	"github.com/iotatfan/hobby-collection-be/internal/text"
	"gorm.io/gorm"
)

func TestParseError_DBErrorValue_RecordNotFound(t *testing.T) {
	msg, code := ParseError(DBError{ErrorMsg: gorm.ErrRecordNotFound})

	if msg != "record not found" {
		t.Fatalf("unexpected message: got %q", msg)
	}
	if code != http.StatusNotFound {
		t.Fatalf("unexpected code: got %d", code)
	}
}

func TestParseError_DBErrorPointer_RecordNotFound(t *testing.T) {
	err := &DBError{ErrorMsg: gorm.ErrRecordNotFound}
	msg, code := ParseError(err)

	if msg != "record not found" {
		t.Fatalf("unexpected message: got %q", msg)
	}
	if code != http.StatusNotFound {
		t.Fatalf("unexpected code: got %d", code)
	}
}

func TestParseError_DBErrorValue_InternalServerError(t *testing.T) {
	msg, code := ParseError(DBError{ErrorMsg: errors.New("db down")})

	if msg != text.ErrServer {
		t.Fatalf("unexpected message: got %q", msg)
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("unexpected code: got %d", code)
	}
}
