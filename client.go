package tezos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-tezos/" + libraryVersion
	mediaType      = "application/json"
)

// NewRequest creates a Tezos RPC request.
func (c *RPCClient) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.UserAgent)

	if ctx != nil {
		return req.WithContext(ctx), nil
	}

	return req, nil
}

// RPCClient manages communication with a Tezos RPC server.
type RPCClient struct {
	// Logger
	Logger Logger
	// HTTP client used to communicate with the Tezos node API.
	Client *http.Client
	// Base URL for API requests.
	BaseURL *url.URL
	// User agent name for client.
	UserAgent string
	// Optional callback for metrics.
	RPCStatusCallback func(req *http.Request, status int, duration time.Duration, err error)
	// Optional callback for metrics.
	RPCHeaderCallback func(req *http.Request, resp *http.Response, duration time.Duration)
}

// NewRPCClient returns a new Tezos RPC client.
func NewRPCClient(httpClient *http.Client, baseURL string) (*RPCClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &RPCClient{
		Client:    httpClient,
		BaseURL:   u,
		UserAgent: userAgent,
	}
	return c, nil
}

func (c *RPCClient) log() Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return log.StandardLogger()
}

func (c *RPCClient) handleNormalResponse(ctx context.Context, resp *http.Response, v interface{}) error {
	// Normal return
	typ := reflect.TypeOf(v)

	if typ.Kind() == reflect.Chan {
		// Handle channel
		dumpResponse(c.log(), log.DebugLevel, resp, false)
		dec := json.NewDecoder(resp.Body)

		cases := []reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(v),
			},
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			},
		}

		for {
			chunkVal := reflect.New(typ.Elem())

			if err := dec.Decode(chunkVal.Interface()); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			spewDump(c.log(), log.TraceLevel, chunkVal.Interface())

			cases[0].Send = chunkVal.Elem()
			if chosen, _, _ := reflect.Select(cases); chosen == 1 {
				return ctx.Err()
			}
		}

		return nil
	}

	// Handle single object
	dumpResponse(c.log(), log.DebugLevel, resp, true)
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&v); err != nil {
		return err
	}

	spewDump(c.log(), log.TraceLevel, v)

	return nil
}

// Do retrieves values from the API and marshals them into the provided interface.
func (c *RPCClient) Do(req *http.Request, v interface{}) (err error) {
	timestamp := time.Now()

	dumpRequest(c.log(), log.DebugLevel, req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	if c.RPCHeaderCallback != nil {
		duration := time.Since(timestamp)
		c.RPCHeaderCallback(req, resp, duration)
	}

	if c.RPCStatusCallback != nil {
		defer func() {
			duration := time.Since(timestamp)
			c.RPCStatusCallback(req, resp.StatusCode, duration, err)
		}()
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	statusClass := resp.StatusCode / 100
	if statusClass == 2 {
		if v == nil {
			return nil
		}
		return c.handleNormalResponse(req.Context(), resp, v)
	}

	// Handle errors
	dumpResponse(c.log(), log.DebugLevel, resp, true)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	httpErr := httpError{
		status:     resp.Status,
		statusCode: resp.StatusCode,
		body:       body,
	}

	if statusClass != 5 || !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		// Other errors with unknown body format (usually human readable string)
		return &httpErr
	}

	var errs Errors
	if err := json.Unmarshal(body, &errs); err != nil {
		return &plainError{&httpErr, fmt.Sprintf("tezos: error decoding RPC error: %v", err)}
	}

	if len(errs) == 0 {
		return &plainError{&httpErr, "tezos: empty error response"}
	}

	return &rpcError{
		httpError: &httpErr,
		errors:    errs,
	}
}
