package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	h       http.Client
	baseUrl *url.URL
	headers map[string]string
}

type ClientOpt func(*Client)

func NewClient(opts ...ClientOpt) *Client {
	c := &Client{
		h:       http.Client{},
		baseUrl: nil,
		headers: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithTransport(tsp http.RoundTripper) ClientOpt {
	return func(c *Client) {
		c.h.Transport = tsp
	}
}

func WithBaseURL(baseUrl string) ClientOpt {
	u, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}
	return func(c *Client) {
		c.baseUrl = u
	}
}

func WithHeader(key, value string) ClientOpt {
	return func(c *Client) {
		c.headers[key] = value
	}
}

func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.h.Do(req)
}

func (c *Client) Post(ctx context.Context, path, contentType string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.h.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.h.Do(req)
}

func (c *Client) NewRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	u, err := c.joinPath(path)
	if err != nil {
		return nil, fmt.Errorf("could not join path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (c *Client) joinPath(path string) (string, error) {
	pu, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if c.baseUrl == nil {
		return pu.String(), nil
	}

	pu.Scheme = c.baseUrl.Scheme
	pu.Host = c.baseUrl.Host
	p, err := url.JoinPath(c.baseUrl.Path, pu.Path)
	if err != nil {
		return "", err
	}
	pu.Path = p
	return pu.String(), nil
}
