package s3api

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strconv"
)

const s3Namespace = "http://s3.amazonaws.com/doc/2006-03-01/"

const timeFormat = "2006-01-02T15:04:05.000Z"

var defaultOwner = owner{ID: "minis3", DisplayName: "minis3"}

type owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type listAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Xmlns   string   `xml:"xmlns,attr"`
	Owner   owner    `xml:"Owner"`
	Buckets struct {
		Bucket []bucketEntry `xml:"Bucket"`
	} `xml:"Buckets"`
}

type bucketEntry struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

func writeXML(w http.ResponseWriter, status int, v any) {
	body, err := xml.Marshal(v)
	if err != nil {
		log.Printf("marshal xml: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("Content-Length", strconv.Itoa(len(xml.Header)+len(body)))
	w.WriteHeader(status)
	io.WriteString(w, xml.Header)
	w.Write(body)
}
