package tezos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-tezos/" + libraryVersion
	mediaType      = "application/json"
)

// RPCErrorKind models the kind of Tezos RPC errors that exist.
type RPCErrorKind int

const (
	// Permanent Tezos RPC error kind.
	Permanent RPCErrorKind = iota
	// Temporary Tezos RPC error kind.
	Temporary
	// Branch Tezos RPC error kind.
	Branch
	// Unknown Tezos RPC error kind.
	Unknown
)

// UnmarshalJSON implements json.Unmarshaler.
func (k *RPCErrorKind) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"permanent"`:
		*k = Permanent
	case `"temporary"`:
		*k = Temporary
	case `"branch"`:
		*k = Branch
	default:
		*k = Unknown
	}
	return nil
}

func (k *RPCErrorKind) String() string {
	switch *k {
	case Permanent:
		return "permanent"
	case Temporary:
		return "temporary"
	case Branch:
		return "branch"
	default:
		return "unknown"
	}
}

// RPCError is a Tezos RPC error as documented on http://tezos.gitlab.io/mainnet/api/errors.html.
type RPCError struct {
	Kind RPCErrorKind `json:"kind"`
	ID   string       `json:"id"`
}

func (k *RPCError) Error() string {
	return fmt.Sprintf("Tezos RPC error (kind = %q, id = %q)", k.Kind, k.ID)
}

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

// Get retrieves values from the API and marshals them into the provided interface.
func (c *RPCClient) Get(ctx context.Context, req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	switch resp.StatusCode / 100 {
	case 4:
		return fmt.Errorf("bad request: %s", resp.Status)
	case 5:
		// Attempt to parse 5xx errors according to http://tezos.gitlab.io/mainnet/api/errors.html.
		var rpcErr RPCError
		return json.NewDecoder(resp.Body).Decode(&rpcErr)
	case 2:
		return json.NewDecoder(resp.Body).Decode(&v)
	default:
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}
}
