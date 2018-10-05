package tezos

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-tezos/" + libraryVersion
	mediaType      = "application/json"
)

// Client manages communication with a Tezos RPC
type Client struct {
	//HTTP client used to comminicate with the DO API.
	client *http.Client

	// Base URL for API requests
	BaseURL *url.URL

	// User agent for clietn
	UserAgent string

	Network NetworkService
	//TODO
	// Injection InjectionService
}

// NewRequest creates a Tezos RPC request
func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
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

// NewClient returns a new Tezos RPC client
func NewClient(httpClient *http.Client, baseURL string) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{client: httpClient, BaseURL: u, UserAgent: userAgent}
	c.Network = &NetworkServiceOp{client: c}
	return c, nil
}

// Do sends an API request and returns an API response
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()
	err = json.NewDecoder(resp.Body).Decode(&v)

	// TODO: check HTTP response codes and handle
	return resp, err
}
