package transit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	baseURL   = "https://api.winnipegtransit.com/v3/"
	userAgent = "winnipeg-transit-go"
)

// A Client manages communication with the Winnipeg Transit API.
type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	APIKey    string

	common service

	Stops *StopsService
}

type service struct {
	client *Client
}

// NewClient returns a new Winnipeg Transit API client for a given apiKey.
// Users can register for an API key at https://api.winnipegtransit.com/home/users/new
func NewClient(apiKey string) *Client {
	parsedBaseURL, _ := url.Parse(baseURL)

	c := &Client{
		client:    &http.Client{},
		BaseURL:   parsedBaseURL,
		UserAgent: userAgent,
		APIKey:    apiKey,
	}

	c.common.client = c
	c.Stops = (*StopsService)(&c.common)
	return c
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
//
// The provided ctx must be non-nil, if it is nil an error is returned.
func (c *Client) NewRequest(ctx context.Context, method, urlStr string) (*http.Request, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}

	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	url, err := c.BaseURL.Parse(urlStr + ".json")
	if err != nil {
		return nil, err
	}

	var buffer io.ReadWriter

	request, err := http.NewRequest(method, url.String(), buffer)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if c.UserAgent != "" {
		request.Header.Set("User-Agent", c.UserAgent)
	}
	return request, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
//
// If the request context is canceled or times out, ctx.Err() will be returned.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		var e *url.Error
		if errors.As(err, &e) {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(url).String()
				return nil, e
			}
		}

		return nil, err
	}

	// I have no idea how this works.
	defer func() {
		// Ensure the response body is fully read and closed
		// before we reconnect, so that we reuse the same TCP connection.
		// Close the previous response's body. But read at least some of
		// the body so if it's small the underlying TCP connection will be
		// re-used. No need to check for errors: if it fails, the Transport
		// won't reuse it anyway.
		const maxBodySlurpSize = 2 << 10
		if resp.ContentLength == -1 || resp.ContentLength <= maxBodySlurpSize {
			io.CopyN(ioutil.Discard, resp.Body, maxBodySlurpSize)
		}

		resp.Body.Close()
	}()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

/*
An ErrorResponse reports one or more errors caused by an API request.
*/
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
	Message  string         `json:"responseText"` // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, sanitizeURL(r.Response.Request.URL),
		r.Response.StatusCode, r.Message)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

// sanitizeURL redacts the api-key parameter from the URL which may be
// exposed to the user.
func sanitizeURL(uri *url.URL) *url.URL {
	if uri == nil {
		return nil
	}
	params := uri.Query()
	if len(params.Get("api-key")) > 0 {
		params.Set("api-key", "REDACTED")
		uri.RawQuery = params.Encode()
	}
	return uri
}
