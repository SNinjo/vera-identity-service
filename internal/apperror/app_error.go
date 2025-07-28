package apperror

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Response struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
type AppError struct {
	Status int `json:"-"`
	Response
}

func (e *AppError) Error() string {
	timestamp := e.Timestamp.Format(time.RFC3339)
	return fmt.Sprintf("AppError{code: \"%s\", message: \"%s\", timestamp: \"%s\"}", e.Code, e.Message, timestamp)
}

var regexpErrorCode = regexp.MustCompile(`^(\d{3})_\d{2}_\d{3}$`)

func New(code string, message string) *AppError {
	parsedStatus := 0
	if matches := regexpErrorCode.FindStringSubmatch(code); len(matches) == 2 {
		parsedStatus, _ = strconv.Atoi(matches[1])
	} else {
		return &AppError{
			Status: 500,
			Response: Response{
				Code:      "config_error",
				Message:   "Invalid the error code format: " + code,
				Timestamp: time.Now().UTC(),
			},
		}
	}
	return &AppError{
		Status: parsedStatus,
		Response: Response{
			Code:      code,
			Message:   message,
			Timestamp: time.Now().UTC(),
		},
	}
}

func FromError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	} else {
		return &AppError{
			Status: 500,
			Response: Response{
				Code:      "unknown_error",
				Message:   err.Error(),
				Timestamp: time.Now().UTC(),
			},
		}
	}
}
