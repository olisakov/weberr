package weberr

import (
	"fmt"
	"io"
	"testing"
)

// TestGetStackTrace tests that the stack always starts with the root cause
func TestGetStackTrace(t *testing.T) {
	fmt.Printf("trace:%s\n", GetStackTrace(nil))

	err := NoType.UserErrorf("user error")
	fmt.Printf("trace:%s\n", GetStackTrace(err))

	err = UserWrapf(err, "wrap user")
	fmt.Printf("error:%v\nuser:%s\ntrace:%s\n", err, GetUserMessage(err), GetStackTrace(err))

	err = Wrapf(err, "Wrapf")
	fmt.Printf("error:%v\nuser:%s\ntrace:%s\n", err, GetUserMessage(err), GetStackTrace(err))

	err1 := NoType.Errorf("error")
	fmt.Printf("trace:%s\n", GetStackTrace(err1))

	err1 = Wrapf(err1, "Wrapf")
	fmt.Printf("error:%v\nuser:%s\ntrace:%s\n", err1, GetUserMessage(err1), GetStackTrace(err1))

	err2 := Wrapf(io.EOF, "End of file")
	fmt.Printf("error:%v\nuser:%s\ntrace:%s\n", err2, GetUserMessage(err2), GetStackTrace(err2))

	err3 := UserWrapf(io.EOF, "End of file")
	fmt.Printf("error:%v\nuser:%s\ntrace:%s\n", err3, GetUserMessage(err3), GetStackTrace(err3))
}

// User message test logic
// GetUser error will return "" for error type that don't implement userMessager interface
// Wrapping in form: "external: internal"
// Wrapf doesn't change the GetUserMessage() message
// UserErrorf sets Error() description initially
func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{nil, ""},
		{io.EOF, ""},
		{UserWrapf(io.EOF, "End Of File"), "End Of File"},
		{NoType.UserErrorf("New User Message!"), "New User Message!"},
		{UserWrapf(BadRequest.UserErrorf("Internal"), "External"), "External: Internal"},
		{UserWrapf(UserWrapf(UserWrapf(io.EOF, "%d", 1), "%s", "2"), "3"), "3: 2: 1"},
	}

	for _, tt := range tests {
		got := GetUserMessage(tt.err)
		if got != tt.expected {
			t.Errorf("got: %q, want %q", got, tt.expected)
		}
	}

}

// Wrapf Test logic
// Wrapped errors are displayed using Error()
// Wrapping in form: "external: internal"
// UserWrapf doesn't change the Error() message
// UserErrorf sets Error() description initially
func TestWrapf(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{io.EOF, "EOF"},
		{Wrapf(io.EOF, "End Of File"), "End Of File: EOF"},
		{NoType.Errorf("New Error!"), "New Error!"},
		{Wrapf(BadRequest.Errorf("Internal"), "External"), "External: Internal"},
		{UserWrapf(BadRequest.Errorf("Internal"), "External"), "Internal"},
		{Wrapf(BadRequest.UserErrorf("Internal"), "External"), "External: Internal"},
		{Wrapf(Unauthorized.Wrapf(Wrapf(io.EOF, "%d", 1), "%s", "2"), "3"), "3: 2: 1: EOF"},
	}

	for _, tt := range tests {
		got := (tt.err.Error())
		if got != tt.expected {
			t.Errorf("got: %q, want %q", got, tt.expected)
		}
	}

}

// Type logic tested:
// Default: NoType
// Type is overridden by top most layer type
// NoType doesn't override an existing type
func TestGetType(t *testing.T) {
	tests := []struct {
		err      error
		expected ErrorType
	}{
		{nil, NoType},
		{io.EOF, NoType},
		{Wrapf(io.EOF, "msg"), NoType},
		{UserWrapf(io.EOF, "msg"), NoType},
		{Errorf("msg"), NoType},
		{UserErrorf("msg"), NoType},
		{BadRequest.UserWrapf(io.EOF, "msg"), BadRequest},
		{UserWrapf(NotFound.UserErrorf("Not Found!"), "a"), NotFound},
		{Wrapf(BadRequest.UserWrapf(io.EOF, "msg"), "no type"), BadRequest},
		{NotFound.Wrapf(Wrapf(BadRequest.UserWrapf(io.EOF, "msg"), "no type"), "Not Found"), NotFound},
		{NotFound.Set(BadRequest.Set(io.EOF)), NotFound},
	}
	for _, tt := range tests {
		got := GetType(tt.err)
		if got != tt.expected {
			t.Errorf("got: %v, want %v", got, tt.expected)
		}
	}

}
