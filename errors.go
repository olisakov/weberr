/*
Copyright (c) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package weberr

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrorType is the type of an error
type ErrorType uint

const (
	// NoType error - HTTP internal Error - code 500
	NoType ErrorType = iota
	// BadRequest error - Code 400
	BadRequest
	// NotFound error - Code 404
	NotFound
	// Unauthorized error - Code 401
	Unauthorized
	// Conflict - Code 409
	Conflict
)

// customError wraps an error with type and user message
type customError struct {
	error
	errorType   ErrorType
	userMessage string
}

// causer interface allows unwrapping an error.
// causer is also used in github.com/pkg/errors
type causer interface {
	Cause() error
}

// Cause unwrappes error
func (c *customError) Cause() error { return c.error }

// typed interface identifies error with a type
type typed interface {
	Type() ErrorType
}

// Type returns the error type
func (c *customError) Type() ErrorType { return c.errorType }

// GetType returns the error type for all errors
// if error is not `typed` - it returns NoType
func GetType(err error) ErrorType {
	if typeErr, ok := err.(typed); ok {
		return typeErr.Type()
	}

	return NoType
}

// userMessager identifies an error with a user message
type userMessager interface {
	UserMessage() string
}

// UserMesage returns the user message
func (c *customError) UserMessage() string { return c.userMessage }

// GetUserMessage returns user readable error message for all errors
// If error is not `userMessager` returns empty string
func GetUserMessage(err error) string {
	if msgErr, ok := err.(userMessager); ok {
		return msgErr.UserMessage()
	}

	return ""
}

// Errorf creates a new customError with formatted message
func (errorType ErrorType) Errorf(msg string, args ...interface{}) error {
	return &customError{
		error:     errors.WithStack(errors.Errorf(msg, args...)),
		errorType: errorType,
	}
}

// Wrapf creates a new wrapped error with formatted message
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	c := new(customError)
	c.error = errors.Wrapf(err, msg, args...)
	c.userMessage = GetUserMessage(err)

	if errorType != NoType {
		c.errorType = errorType
	} else {
		c.errorType = GetType(err)
	}

	return c
}

// UserWrapf adds a user readable to an error
func (errorType ErrorType) UserWrapf(err error, msg string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	userMsg := fmt.Sprintf(msg, args...)

	c := new(customError)
	c.error = errors.WithStack(err)

	origMsg := GetUserMessage(err)
	if origMsg != "" {
		userMsg = fmt.Sprintf("%s: %s", userMsg, origMsg)
	}
	c.userMessage = userMsg

	if errorType != NoType {
		c.errorType = errorType
	} else {
		c.errorType = GetType(err)
	}

	return c

}

// UserErrorf creates a new error with a user message
func (errorType ErrorType) UserErrorf(msg string, args ...interface{}) error {
	message := fmt.Sprintf(msg, args...)
	return &customError{
		error:       errors.WithStack(errors.New(message)),
		errorType:   errorType,
		userMessage: message,
	}
}

// Set the type of the error
func (errorType ErrorType) Set(err error) error {
	if err == nil {
		return nil
	}

	return &customError{
		error:       errors.WithStack(err),
		errorType:   errorType,
		userMessage: GetUserMessage(err),
	}
}

// Errorf returns an error with format string
func Errorf(msg string, args ...interface{}) error {
	return NoType.Errorf(msg, args...)
}

// Wrapf return an error with format string
func Wrapf(err error, msg string, args ...interface{}) error {
	return NoType.Wrapf(err, msg, args...)
}

// UserErrorf returns an error with format string
func UserErrorf(msg string, args ...interface{}) error {
	return NoType.UserErrorf(msg, args...)
}

// UserWrapf adds a user readable to an error
func UserWrapf(err error, msg string, args ...interface{}) error {
	return NoType.UserWrapf(err, msg, args...)
}

// stackTracer interface is internally defined in github.com/pkg/errors
// and identifies an error with a stack trace
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// baseStackTracer is a helper function to allow reaching
// the initial wrapper that has a stack trace
func baseStackTracer(err error) error {

	if cause, ok := err.(causer); ok {
		candidate := baseStackTracer(cause.Cause())
		if candidate != nil {
			return candidate
		}

		if _, ok := err.(stackTracer); ok {
			return err
		}
	}
	return nil
}

// GetStackTrace returns the stack trace starting from the first error
// that has been wrapped / created
func GetStackTrace(err error) string {
	if err == nil {
		return ""
	}

	err = baseStackTracer(err)
	x, ok := err.(stackTracer)
	if !ok {
		// The error doen't have a stack trace attached to it
		return fmt.Sprintf("%+v", err)
	}

	st := x.StackTrace()
	return fmt.Sprintf("%+v\n", st[1:])
}