package p2jsvr

import "encoding/json"

type HttpError struct {
	message string
	code    int
}

func NewHttpError(msg string, code int) *HttpError {
	return &HttpError{
		message: msg,
		code:    code,
	}
}

func (e *HttpError) Code() int {
	return e.code
}

func (e *HttpError) JsonSerialize() string {
	type resp struct {
		Message string `json:"message"`
	}
	str, _ := json.Marshal(resp{Message: e.message})
	return string(str)
}
