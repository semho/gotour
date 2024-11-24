// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: kafka/v1/kafka.proto

package kafka_v1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on ChatMessageEvent with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *ChatMessageEvent) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ChatMessageEvent with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ChatMessageEventMultiError, or nil if none found.
func (m *ChatMessageEvent) ValidateAll() error {
	return m.validate(true)
}

func (m *ChatMessageEvent) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetMetadata()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ChatMessageEventValidationError{
					field:  "Metadata",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ChatMessageEventValidationError{
					field:  "Metadata",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetMetadata()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ChatMessageEventValidationError{
				field:  "Metadata",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if all {
		switch v := interface{}(m.GetPayload()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ChatMessageEventValidationError{
					field:  "Payload",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ChatMessageEventValidationError{
					field:  "Payload",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetPayload()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ChatMessageEventValidationError{
				field:  "Payload",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return ChatMessageEventMultiError(errors)
	}

	return nil
}

// ChatMessageEventMultiError is an error wrapping multiple validation errors
// returned by ChatMessageEvent.ValidateAll() if the designated constraints
// aren't met.
type ChatMessageEventMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ChatMessageEventMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ChatMessageEventMultiError) AllErrors() []error { return m }

// ChatMessageEventValidationError is the validation error returned by
// ChatMessageEvent.Validate if the designated constraints aren't met.
type ChatMessageEventValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ChatMessageEventValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ChatMessageEventValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ChatMessageEventValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ChatMessageEventValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ChatMessageEventValidationError) ErrorName() string { return "ChatMessageEventValidationError" }

// Error satisfies the builtin error interface
func (e ChatMessageEventValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sChatMessageEvent.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ChatMessageEventValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ChatMessageEventValidationError{}

// Validate checks the field values on ChatMessageEvent_Metadata with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ChatMessageEvent_Metadata) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ChatMessageEvent_Metadata with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ChatMessageEvent_MetadataMultiError, or nil if none found.
func (m *ChatMessageEvent_Metadata) ValidateAll() error {
	return m.validate(true)
}

func (m *ChatMessageEvent_Metadata) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for EventId

	if all {
		switch v := interface{}(m.GetCreatedAt()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ChatMessageEvent_MetadataValidationError{
					field:  "CreatedAt",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ChatMessageEvent_MetadataValidationError{
					field:  "CreatedAt",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetCreatedAt()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ChatMessageEvent_MetadataValidationError{
				field:  "CreatedAt",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for EventType

	if len(errors) > 0 {
		return ChatMessageEvent_MetadataMultiError(errors)
	}

	return nil
}

// ChatMessageEvent_MetadataMultiError is an error wrapping multiple validation
// errors returned by ChatMessageEvent_Metadata.ValidateAll() if the
// designated constraints aren't met.
type ChatMessageEvent_MetadataMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ChatMessageEvent_MetadataMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ChatMessageEvent_MetadataMultiError) AllErrors() []error { return m }

// ChatMessageEvent_MetadataValidationError is the validation error returned by
// ChatMessageEvent_Metadata.Validate if the designated constraints aren't met.
type ChatMessageEvent_MetadataValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ChatMessageEvent_MetadataValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ChatMessageEvent_MetadataValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ChatMessageEvent_MetadataValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ChatMessageEvent_MetadataValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ChatMessageEvent_MetadataValidationError) ErrorName() string {
	return "ChatMessageEvent_MetadataValidationError"
}

// Error satisfies the builtin error interface
func (e ChatMessageEvent_MetadataValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sChatMessageEvent_Metadata.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ChatMessageEvent_MetadataValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ChatMessageEvent_MetadataValidationError{}

// Validate checks the field values on ChatMessageEvent_Payload with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *ChatMessageEvent_Payload) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ChatMessageEvent_Payload with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ChatMessageEvent_PayloadMultiError, or nil if none found.
func (m *ChatMessageEvent_Payload) ValidateAll() error {
	return m.validate(true)
}

func (m *ChatMessageEvent_Payload) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for MessageId

	// no validation rules for ChatId

	// no validation rules for SessionId

	// no validation rules for Nickname

	// no validation rules for Text

	if all {
		switch v := interface{}(m.GetTimestamp()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ChatMessageEvent_PayloadValidationError{
					field:  "Timestamp",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ChatMessageEvent_PayloadValidationError{
					field:  "Timestamp",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetTimestamp()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ChatMessageEvent_PayloadValidationError{
				field:  "Timestamp",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return ChatMessageEvent_PayloadMultiError(errors)
	}

	return nil
}

// ChatMessageEvent_PayloadMultiError is an error wrapping multiple validation
// errors returned by ChatMessageEvent_Payload.ValidateAll() if the designated
// constraints aren't met.
type ChatMessageEvent_PayloadMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ChatMessageEvent_PayloadMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ChatMessageEvent_PayloadMultiError) AllErrors() []error { return m }

// ChatMessageEvent_PayloadValidationError is the validation error returned by
// ChatMessageEvent_Payload.Validate if the designated constraints aren't met.
type ChatMessageEvent_PayloadValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ChatMessageEvent_PayloadValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ChatMessageEvent_PayloadValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ChatMessageEvent_PayloadValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ChatMessageEvent_PayloadValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ChatMessageEvent_PayloadValidationError) ErrorName() string {
	return "ChatMessageEvent_PayloadValidationError"
}

// Error satisfies the builtin error interface
func (e ChatMessageEvent_PayloadValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sChatMessageEvent_Payload.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ChatMessageEvent_PayloadValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ChatMessageEvent_PayloadValidationError{}
