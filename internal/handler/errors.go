package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jaysongiroux/mdserve/internal/logger"
)

type PageError struct {
	Code    string
	Title   string
	Message string
}

func (e *PageError) Error() string {
	return fmt.Sprintf("%s: %s - %s", e.Code, e.Title, e.Message)
}

func NewPageError(code, title, message string) *PageError {
	return &PageError{
		Code:    code,
		Title:   title,
		Message: message,
	}
}

func handleError(app *App, w http.ResponseWriter, err error, data *TemplateData) {
	pageErr := &PageError{}
	if errors.As(err, &pageErr) {
		data.ErrorCode = &pageErr.Code
		data.ErrorTitle = &pageErr.Title
		data.ErrorMessage = &pageErr.Message
	} else {
		data.ErrorCode = &Err500Code
		data.ErrorTitle = &Err500Title
		data.ErrorMessage = &Err500Message
	}
	err = app.Templates.ExecuteTemplate(w, "error.html", data)
	if err != nil {
		logger.Error("Failed to execute template: %v", err)
	}
}
