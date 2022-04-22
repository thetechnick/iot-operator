package clients

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

	"github.com/thetechnick/iot-operator/internal/version"
)

type Client struct {
	opts       ClientOptions
	apiErrType reflect.Type
	httpClient *http.Client
}

type ClientOptions struct {
	Endpoint  string
	ErrorType error
}

type ClientOption interface {
	ApplyToClient(o *ClientOptions)
}

type WithEndpoint string

func (e WithEndpoint) ApplyToClient(o *ClientOptions) {
	// ensure there is always a single trailing "/"
	o.Endpoint = strings.TrimRight(string(e), "/") + "/"
}

type WithAPIErrType struct{ APIError error }

func (e WithAPIErrType) ApplyToClient(o *ClientOptions) {
	o.ErrorType = e.APIError
}

// Creates a new OCM client with the given options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt.ApplyToClient(&c.opts)
	}

	c.apiErrType = reflect.TypeOf(c.opts.ErrorType)
	c.httpClient = &http.Client{}
	return c
}

func (c *Client) Do(
	ctx context.Context,
	httpMethod string,
	path string,
	params url.Values,
	payload, result interface{},
) error {
	// Build URL
	reqURL, err := url.Parse(c.opts.Endpoint)
	if err != nil {
		return fmt.Errorf("parsing endpoint URL: %w", err)
	}
	reqURL = reqURL.ResolveReference(&url.URL{
		Path: strings.TrimLeft(path, "/"), // trim first slash to always be relative to baseURL
	})

	// Payload
	var resBody io.Reader
	if payload != nil {
		j, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshaling json: %w", err)
		}

		resBody = bytes.NewBuffer(j)
	}

	var fullUrl string
	if len(params) > 0 {
		fullUrl = reqURL.String() + "?" + params.Encode()
	} else {
		fullUrl = reqURL.String()
	}

	httpReq, err := http.NewRequestWithContext(ctx, httpMethod, fullUrl, resBody)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}

	// Headers
	httpReq.Header.Add("User-Agent", fmt.Sprintf("IoTOperator/%s", version.Version))
	httpReq.Header.Add("Content-Type", "application/json")

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("executing http request: %w", err)
	}
	defer httpRes.Body.Close()

	// HTTP Error handling
	if httpRes.StatusCode >= 400 && httpRes.StatusCode <= 599 {
		body, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return fmt.Errorf("reading error response body %s: %w", fullUrl, err)
		}

		if c.apiErrType == nil {
			return fmt.Errorf("HTTP %d: %s", httpRes.StatusCode, body)
		}

		apiErr := reflect.New(c.apiErrType).Interface()
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf(
				"HTTP %d: unmarshal json error response %s: %w", httpRes.StatusCode, string(body), err)
		}
		return apiErr.(error)
	}

	// Read response
	if result != nil {
		body, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return fmt.Errorf("reading response body %s: %w", fullUrl, err)
		}

		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("unmarshal json response %s: %w", fullUrl, err)
		}
	}

	return nil
}
