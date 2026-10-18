package response

import "net/http"

// Response describes a transport result before it is written to the connection.
type Response struct {
	Status      int
	Payload     any
	Headers     http.Header
	Cookies     []*http.Cookie
	InternalErr error
}

func (r *Response) WithHeaders(headers map[string]string) *Response {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}

	for key, value := range headers {
		r.Headers.Set(key, value)
	}

	return r
}

func (r *Response) WithCookies(cookies ...*http.Cookie) *Response {
	r.Cookies = append(r.Cookies, cookies...)

	return r
}

func OK(payload any) *Response       { return &Response{Status: http.StatusOK, Payload: payload} }
func Created(payload any) *Response  { return &Response{Status: http.StatusCreated, Payload: payload} }
func Accepted(payload any) *Response { return &Response{Status: http.StatusAccepted, Payload: payload} }
func NoContent() *Response           { return &Response{Status: http.StatusNoContent} }
