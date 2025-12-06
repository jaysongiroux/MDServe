package handler

import (
	"fmt"
	"net/http"
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
	if pageErr, ok := err.(*PageError); ok {
		data.ErrorCode = &pageErr.Code
		data.ErrorTitle = &pageErr.Title
		data.ErrorMessage = &pageErr.Message
	} else {
		data.ErrorCode = &Err500Code
		data.ErrorTitle = &Err500Title
		data.ErrorMessage = &Err500Message
	}
	app.Templates.ExecuteTemplate(w, "error.html", data)
}
