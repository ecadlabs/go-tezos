package tezos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

// Logger is an extension to logrus.FieldLogger
type Logger interface {
	log.FieldLogger
	Writer() *io.PipeWriter
	WriterLevel(level log.Level) *io.PipeWriter
}

/*
unmarshalHeterogeneousJSONArray is a helper function used in custom JSON
unmarshallers and intended to decode array-like objects:
	[
		"...", // object ID or hash
		{
			... // ebject with ID ommitted
		}
	]
*/
func unmarshalHeterogeneousJSONArray(data []byte, v ...interface{}) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) < len(v) {
		return fmt.Errorf("JSON array is too short, expected %d, got %d", len(v), len(raw))
	}

	for i, vv := range v {
		if err := json.Unmarshal(raw[i], vv); err != nil {
			return err
		}
	}

	return nil
}

// unmarshalInSlice unmarshals a JSON array in a way so that each element of the
// interface slice is unmarshaled individually. This is a workaround for the
// case where Go's normal unmarshaling wants to treat the array as a whole.
func unmarshalInSlice(data []byte, s []interface{}) error {
	var aRaw []json.RawMessage
	if err := json.Unmarshal(data, &aRaw); err != nil {
		return err
	}

	if len(aRaw) != len(s) {
		return fmt.Errorf("Array is too short, JSON has %d, we have %d", len(aRaw), len(s))
	}

	for i, raw := range aRaw {
		if err := json.Unmarshal(raw, &s[i]); err != nil {
			return err
		}
	}
	return nil
}

func isLevelEnabled(logger Logger, level log.Level) bool {
	switch l := logger.(type) {
	case *log.Entry:
		return l.Logger.IsLevelEnabled(level)
	case *log.Logger:
		return l.IsLevelEnabled(level)
	}
	return false
}

func spewDump(logger Logger, level log.Level, v ...interface{}) {
	if !isLevelEnabled(logger, level) {
		return
	}

	w := logger.WriterLevel(level)
	defer w.Close()

	spew.Fdump(w, v...)
}

func dumpRequest(logger Logger, level log.Level, req *http.Request) {
	if !isLevelEnabled(logger, level) {
		return
	}

	buf, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Error(err)
		return
	}

	w := logger.WriterLevel(level)
	defer w.Close()

	w.Write(buf)
}

func dumpResponse(logger Logger, level log.Level, res *http.Response, body bool) {
	if !isLevelEnabled(logger, level) {
		return
	}

	buf, err := httputil.DumpResponse(res, body)
	if err != nil {
		logger.Error(err)
		return
	}

	w := logger.WriterLevel(level)
	defer w.Close()

	w.Write(buf)
}
