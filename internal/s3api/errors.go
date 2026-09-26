package s3api

import (
	"encoding/xml"
	"errors"
	"log"
	"net/http"

	"github.com/ihsanguldur/mini-s3/internal/metadata"
)

type s3Error struct {
	Code    string
	Message string
	Status  int
}

func (e *s3Error) Error() string {
	return e.Code + ": " + e.Message
}

var (
	errNoSuchBucket = &s3Error{"NoSuchBucket",
		"The specified bucket does not exist.", http.StatusNotFound}
	errBucketAlreadyOwnedByYou = &s3Error{"BucketAlreadyOwnedByYou",
		"Your previous request to create the named bucket succeeded and you already own it.", http.StatusConflict}
	errInvalidBucketName = &s3Error{"InvalidBucketName",
		"The specified bucket is not valid.", http.StatusBadRequest}
	errNotImplemented = &s3Error{"NotImplemented",
		"A header or query you provided implies functionality that is not implemented.", http.StatusNotImplemented}
	errInternal = &s3Error{"InternalError",
		"We encountered an internal error. Please try again.", http.StatusInternalServerError}
)

type errorResponse struct {
	XMLName  xml.Name `xml:"Error"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
	Resource string   `xml:"Resource"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var s3err *s3Error
	switch {
	case errors.As(err, &s3err):
	case errors.Is(err, metadata.ErrBucketNotFound):
		s3err = errNoSuchBucket
	case errors.Is(err, metadata.ErrBucketExists):
		s3err = errBucketAlreadyOwnedByYou
	default:
		log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
		s3err = errInternal
	}
	writeXML(w, s3err.Status, errorResponse{
		Code:     s3err.Code,
		Message:  s3err.Message,
		Resource: r.URL.Path,
	})
}
