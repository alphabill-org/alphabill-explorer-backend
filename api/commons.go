package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/fxamacker/cbor/v2"
)

const (
	ContentType     = "Content-Type"
	ApplicationJson = "application/json"
	ApplicationCbor = "application/cbor"
	UserAgent       = "User-Agent"

	QueryParamOffsetKey   = "offsetKey"
	QueryParamLimit       = "limit"
	HeaderLink            = "Link"
	HeaderLinkValueFormat = `<%s>; rel="next"`
)

var (
	// ErrInvalidRequest is returned when backend responded with 4nn status code, use errors.Is to check for it.
	ErrInvalidRequest = errors.New("invalid request")

	// ErrNotFound is returned when backend responded with 404 status code, use errors.Is to check for it.
	ErrNotFound = errors.New("not found")
)

type (
	EmptyResponse struct{}

	ErrorResponse struct {
		Message string `json:"message"`
	}

	ResponseWriter struct {
		//LogErr func(err error)
	}
)

func (rw *ResponseWriter) WriteResponse(w http.ResponseWriter, data any) {
	w.Header().Set(ContentType, ApplicationJson)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		//rw.logError(fmt.Errorf("failed to encode response data as json: %w", err))
	}
}

func (rw *ResponseWriter) WriteCborResponse(w http.ResponseWriter, data any) {
	w.Header().Set(ContentType, ApplicationCbor)
	if err := cbor.NewEncoder(w).Encode(data); err != nil {
		//rw.logError(fmt.Errorf("failed to encode response data as cbor: %w", err))
	}
}

func (rw *ResponseWriter) WriteInternalErrorResponse(w http.ResponseWriter, err error) {
	fmt.Printf("Internal error: %s\n", err)
	rw.ErrorResponse(w, http.StatusInternalServerError, errors.New("internal error"))
}

func (rw *ResponseWriter) WriteErrorResponse(w http.ResponseWriter, err error, statusCode ...int) {
	if len(statusCode) > 0 {
		rw.ErrorResponse(w, statusCode[0], err)
		return
	}
	rw.ErrorResponse(w, http.StatusBadRequest, err)
}

func (rw *ResponseWriter) WriteMissingParamResponse(w http.ResponseWriter, param string) {
	rw.ErrorResponse(w, http.StatusBadRequest, fmt.Errorf("missing '%s' parameter", param))
}

func (rw *ResponseWriter) WriteInvalidParamResponse(w http.ResponseWriter, param string) {
	rw.ErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid '%s' parameter", param))
}

func (rw *ResponseWriter) ErrorResponse(w http.ResponseWriter, code int, err error) {
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()}); err != nil {
		//rw.logError(fmt.Errorf("failed to encode error response as json: %w", err))
	}
}

func setLinkHeader(u *url.URL, w http.ResponseWriter, next string) {
	if next == "" {
		w.Header().Del(HeaderLink)
		return
	}
	qp := u.Query()
	qp.Set(QueryParamOffsetKey, next)
	u.RawQuery = qp.Encode()
	w.Header().Set(HeaderLink, fmt.Sprintf(HeaderLinkValueFormat, u))
}

/*
parseMaxResponseItems parses input "s" as integer.
When empty string or int over "maxValue" is sent in "maxValue" is returned.
In case of invalid int or value smaller than 1 error is returned.
*/
func ParseMaxResponseItems(s string, maxValue int) (int, error) {
	if s == "" {
		return maxValue, nil
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %q as integer: %w", s, err)
	}
	if v <= 0 {
		return 0, fmt.Errorf("value must be greater than zero, got %d", v)
	}

	if v > maxValue {
		return maxValue, nil
	}
	return v, nil
}

// DecodeResponse when "rsp" StatusCode is equal to "successStatus" response body is decoded into "data".
// In case of some other response status body is expected to contain error response json struct.
func DecodeResponse(rsp *http.Response, successStatus int, data any, allowEmptyResponse bool) error {
	defer rsp.Body.Close()
	type Decoder interface {
		Decode(val interface{}) error
	}
	var dec Decoder
	contentType := rsp.Header.Get(ContentType)
	if contentType == ApplicationCbor {
		dec = cbor.NewDecoder(rsp.Body)
	} else {
		dec = json.NewDecoder(rsp.Body)
	}
	if rsp.StatusCode == successStatus {
		err := dec.Decode(data)
		if err != nil && (!errors.Is(err, io.EOF) || !allowEmptyResponse) {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
		return nil
	}

	var errResponse ErrorResponse
	if err := json.NewDecoder(rsp.Body).Decode(&errResponse); err != nil {
		return fmt.Errorf("failed to decode error from the response body (%s): %w", rsp.Status, err)
	}

	msg := fmt.Sprintf("backend responded %s: %s", rsp.Status, errResponse.Message)
	switch {
	case rsp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("%s: %w", errResponse.Message, ErrNotFound)
	case 400 <= rsp.StatusCode && rsp.StatusCode < 500:
		return fmt.Errorf("%s: %w", msg, ErrInvalidRequest)
	}
	return errors.New(msg)
}
