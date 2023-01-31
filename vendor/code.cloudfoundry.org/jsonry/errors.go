package jsonry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"code.cloudfoundry.org/jsonry/internal/errorcontext"
)

type unsupportedType struct {
	typ reflect.Type
}

func newUnsupportedTypeError(t reflect.Type) error {
	return &unsupportedType{
		typ: t,
	}
}

func (u unsupportedType) Error() string {
	return u.message(errorcontext.ErrorContext{})
}

func (u unsupportedType) message(ctx errorcontext.ErrorContext) string {
	return fmt.Sprintf(`unsupported type "%s" at %s`, u.typ, ctx)
}

type unsupportedKeyType struct {
	typ reflect.Type
}

func newUnsupportedKeyTypeError(t reflect.Type) error {
	return &unsupportedKeyType{
		typ: t,
	}
}

func (u unsupportedKeyType) Error() string {
	return u.message(errorcontext.ErrorContext{})
}

func (u unsupportedKeyType) message(ctx errorcontext.ErrorContext) string {
	return fmt.Sprintf(`maps must only have string keys for "%s" at %s`, u.typ, ctx)
}

type conversionError struct {
	value interface{}
}

func newConversionError(value interface{}) error {
	return &conversionError{
		value: value,
	}
}

func (c conversionError) Error() string {
	return c.message(errorcontext.ErrorContext{})
}

func (c conversionError) message(ctx errorcontext.ErrorContext) string {
	var t string
	switch c.value.(type) {
	case nil:
	case json.Number:
		t = "number"
	default:
		t = reflect.TypeOf(c.value).String()
	}

	msg := fmt.Sprintf(`cannot unmarshal "%+v" `, c.value)

	if t != "" {
		msg = fmt.Sprintf(`%stype "%s" `, msg, t)
	}

	return msg + "into " + ctx.String()
}

type foreignError struct {
	msg   string
	cause error
}

func newForeignError(msg string, cause error) error {
	return foreignError{
		msg:   msg,
		cause: cause,
	}
}

func (e foreignError) Error() string {
	return e.message(errorcontext.ErrorContext{})
}

func (e foreignError) message(ctx errorcontext.ErrorContext) string {
	return fmt.Sprintf("%s at %s: %s", e.msg, ctx, e.cause)
}

type contextError struct {
	cause   error
	context errorcontext.ErrorContext
}

func (c contextError) Error() string {
	if e, ok := c.cause.(interface {
		message(errorcontext.ErrorContext) string
	}); ok {
		return e.message(c.context)
	}
	return c.cause.Error()
}

func wrapErrorWithFieldContext(err error, fieldName string, fieldType reflect.Type) error {
	switch e := err.(type) {
	case contextError:
		e.context = e.context.WithField(fieldName, fieldType)
		return e
	default:
		return contextError{
			cause:   err,
			context: errorcontext.ErrorContext{}.WithField(fieldName, fieldType),
		}
	}
}

func wrapErrorWithIndexContext(err error, index int, elementType reflect.Type) error {
	switch e := err.(type) {
	case contextError:
		e.context = e.context.WithIndex(index, elementType)
		return e
	default:
		return contextError{
			cause:   err,
			context: errorcontext.ErrorContext{}.WithIndex(index, elementType),
		}
	}
}

func wrapErrorWithKeyContext(err error, keyName string, valueType reflect.Type) error {
	switch e := err.(type) {
	case contextError:
		e.context = e.context.WithKey(keyName, valueType)
		return e
	default:
		return contextError{
			cause:   err,
			context: errorcontext.ErrorContext{}.WithKey(keyName, valueType),
		}
	}
}
