package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/integr8ly/grafana-operator/pkg/api/models"
)

type ErrorResponse interface {
	SetStatusCode(code int)
	SetPayload(payload *models.Error)
	WriteResponse(rw http.ResponseWriter, producer runtime.Producer)
}

func NewErrorResponse(resp ErrorResponse, code int, msg string, a ...interface{}) ErrorResponse {
	resp.SetStatusCode(code)
	m := fmt.Sprintf(msg, a...)
	c := int64(code)

	resp.SetPayload(&models.Error{
		Message: m,
		Code:    c,
	})
	return resp
}
