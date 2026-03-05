package helper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSONResponse is to standardize the response
type JSONResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	IsError bool   `json:"isError"`
}

// SuccessResponse is to return success JSON response
func SuccessResponse(c *gin.Context, data any, statusCode int, message ...string) {
	code := statusCode
	if code == 0 {
		code = http.StatusOK
	}
	response := "success"
	if len(message) > 0 {
		response = message[0]
	}

	returnJSONResponse(c, response, data, code, false)
}

// ErrorResponse is to return error JSON response
func ErrorResponse(c *gin.Context, err error) {
	errMsg, code := ParseError(err)
	returnJSONResponse(c, errMsg, nil, code, true)
}

// returnJSONResponse is a helper function, returning JSON value containing message and data
func returnJSONResponse(c *gin.Context, message string, data any, statusCode int, isError bool) {
	response := JSONResponse{
		Message: message,
		Data:    data,
		IsError: isError,
		Code:    statusCode,
	}
	c.JSON(statusCode, response)
}

func NoRouteHandler(c *gin.Context) {
	ErrorResponse(c, NoRouteError{ErrorMsg: "resource not found"})
}

func NoMethodHandler(c *gin.Context) {
	ErrorResponse(c, NoRouteError{ErrorMsg: "resource not found"})
}
