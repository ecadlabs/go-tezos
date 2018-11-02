package tezos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-tezos/" + libraryVersion
	mediaType      = "application/json"
)

const (
	// ErrorKindPermanent Tezos RPC error kind.
	ErrorKindPermanent = "permanent"
	// ErrorKindTemporary Tezos RPC error kind.
	ErrorKindTemporary = "temporary"
	// ErrorKindBranch Tezos RPC error kind.
	ErrorKindBranch = "branch"
)

// HTTPError retains HTTP status
type HTTPError interface {
	Status() string  // e.g. "200 OK"
	StatusCode() int // e.g. 200
	Body() []byte
}

type httpError struct {
	status     string
	statusCode int
	body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("tezos: HTTP status %v)", e.statusCode)
}

func (e *httpError) Status() string {
	return e.status
}

func (e *httpError) StatusCode() int {
	return e.statusCode
}

func (e *httpError) Body() []byte {
	return e.body
}

// RPCError is a Tezos RPC error as documented on http://tezos.gitlab.io/mainnet/api/errors.html.
type RPCError struct {
	*httpError
	ID   string
	Kind string // e.g. "permanent"
	Raw  map[string]interface{}
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("tezos: RPC error (kind = %q, id = %q)", e.Kind, e.ID)
}

var _ HTTPError = &RPCError{}

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

	return req, nil
}

// RPCClient manages communication with a Tezos RPC server.
type RPCClient struct {
	// HTTP client used to communicate with the Tezos node API.
	client *http.Client
	// Base URL for API requests.
	BaseURL *url.URL
	// User agent name for client.
	UserAgent string
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

	c := &RPCClient{client: httpClient, BaseURL: u, UserAgent: userAgent}
	return c, nil
}

var errErrDecoding = errors.New("tezos: error decoding RPC error")

// Get retrieves values from the API and marshals them into the provided interface.
func (c *RPCClient) Get(ctx context.Context, req *http.Request, v interface{}) (err error) {
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	statusClass := resp.StatusCode / 100
	if statusClass == 2 {
		// Normal return
		return json.NewDecoder(resp.Body).Decode(&v)
	}

	// Handle errors
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	httpErr := httpError{
		status:     resp.Status,
		statusCode: resp.StatusCode,
		body:       body,
	}

	if statusClass != 5 {
		// Other errors with unknown body format (usually human readable string)
		return &httpErr
	}

	var rawError map[string]interface{}
	if err = json.Unmarshal(body, &rawError); err != nil {
		return fmt.Errorf("tezos: error decoding RPC error: %v", err)
	}

	errID, ok := rawError["id"].(string)
	if !ok {
		return errErrDecoding
	}

	errKind, ok := rawError["kind"].(string)
	if !ok {
		return errErrDecoding
	}

	rpcErr := RPCError{
		httpError: &httpErr,
		ID:        errID,
		Kind:      errKind,
		Raw:       rawError,
	}

	return &rpcErr
}
