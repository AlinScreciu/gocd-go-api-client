package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlinScreciu/gocd-go-api-client/internal/logging"
	"github.com/sirupsen/logrus"
)

type AuthType uint8

const (
	None AuthType = iota
	Basic
	AccessToken
)

type Client struct {
	ctx        context.Context
	ServerURL  *url.URL
	HttpClient *http.Client
	Debug      bool
	auth       AuthType
	user       string
	password   string
	token      string
}

func (c *Client) SetBasicAuth(user, password string) {
	c.auth = Basic
	c.user = user
	c.password = password
	c.token = ""
}

func (c *Client) SetAccessToken(token string) {
	c.auth = AccessToken
	c.user = ""
	c.password = ""
	c.token = token
}

func NewClient(ctx context.Context, server *url.URL) *Client {
	return &Client{
		ServerURL: server,
		HttpClient: &http.Client{
			Timeout: time.Minute,
		},
		ctx: ctx,
	}
}

func setAuth(c *Client, req *http.Request) {
	switch c.auth {
	case Basic:
		req.SetBasicAuth(c.user, c.password)
	case AccessToken:
		req.Header.Add("Authorization", "Bearer "+c.token)
	}
}

func Get[T any](c *Client, endpoint, accept, module string) (*T, error) {
	url := c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}
	logger := l.WithFields(logrus.Fields{
		"METHOD": "GET",
		"URL":    url,
	})

	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("failed to create request object, url: %s: '%s'", url, err.Error())

		return nil, fmt.Errorf("failed to create request object, url: %s: '%w'", url, err)
	}

	req.Header.Add("Accept", accept)

	setAuth(c, req)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Errorf("%s", err.Error())

		return nil, fmt.Errorf("request failed, url: %s: '%w'", url, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))

		if err == nil {
			sb.WriteString(fmt.Sprintf(": '%s'", string(body)))
		}

		errMsg := sb.String()

		logger.Error(errMsg)

		return nil, errors.New(errMsg)
	}

	logger.Infof("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	if err != nil {
		logger.Errorf("failed to read response body: '%s'", err.Error())

		return nil, fmt.Errorf("failed to read response body: '%w'", err)
	}

	var t T

	err = json.Unmarshal(body, &t)
	if err != nil {
		logger.Errorf("failed to parse response body: '%s'", err.Error())

		return nil, fmt.Errorf("failed to parse response body: '%w'", err)
	}

	return &t, nil
}

func GetWithETag[T any](c *Client, endpoint, accept, module string) (*T, string, error) {
	url := c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}
	logger := l.WithFields(logrus.Fields{
		"METHOD": "GET",
		"URL":    url,
	})

	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("failed to create request object, url: %s: '%s'", url, err.Error())

		return nil, "", fmt.Errorf("failed to create request object, url: %s: '%w'", url, err)
	}

	req.Header.Add("Accept", accept)

	setAuth(c, req)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Errorf("%s", err.Error())

		return nil, "", fmt.Errorf("request failed, url: %s: '%w'", url, err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))

		if err == nil {
			sb.WriteString(fmt.Sprintf(": '%s'", string(body)))
		}

		errMsg := sb.String()
		logger.Error(errMsg)

		return nil, "", errors.New(errMsg)
	}

	logger.Infof("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	if err != nil {
		logger.Errorf("failed to read response body: '%s'", err.Error())

		return nil, "", fmt.Errorf("failed to read response body: '%w'", err)
	}

	var t T

	err = json.Unmarshal(body, &t)
	if err != nil {
		logger.Errorf("failed to parse response body: '%s'", err.Error())

		return nil, "", fmt.Errorf("failed to parse response body: '%w'", err)
	}

	eTag := res.Header.Get("ETag")
	if eTag == "" {
		return nil, "", errors.New("missing or empty ETag header")
	}

	return &t, eTag, nil
}

func Put[P any, R any](c *Client, payload *P, eTag, endpoint, accept, module string) (*R, error) {
	url := c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}

	logger := l.WithFields(logrus.Fields{
		"METHOD": http.MethodPut,
		"URL":    url,
	})

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	err := enc.Encode(*payload)
	if err != nil {
		logger.Errorf("failed to encode payload: '%s'", err.Error())
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPut, url, &buf)
	if err != nil {
		logger.Errorf("failed to create request object: '%s'", err.Error())

		return nil, fmt.Errorf("failed to create request object: '%w'", err)
	}

	req.Header.Add("Accept", accept)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("If-Match", eTag)

	setAuth(c, req)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return nil, fmt.Errorf("request failed: '%w'", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))

		if err == nil {
			sb.WriteString(fmt.Sprintf(": '%s'", string(body)))
		}

		errMsg := sb.String()

		logger.Error(errMsg)

		return nil, errors.New(errMsg)
	}

	logger.Infof("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	if err != nil {
		logger.Errorf("failed to read response body: '%s'", err.Error())
		return nil, fmt.Errorf("failed to read response body: '%w'", err)
	}

	var r R

	err = json.Unmarshal(body, &r)
	if err != nil {
		logger.Errorf("failed to parse response body: '%s'", err.Error())
		return nil, fmt.Errorf("failed to parse response body: '%w'", err)
	}

	return &r, nil
}

func Post[P any, R any](c *Client, payload *P, endpoint string, accept string, module string) (*R, error) {
	url := c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}

	logger := l.WithFields(logrus.Fields{
		"METHOD": http.MethodPost,
		"URL":    url,
	})

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)

	err := enc.Encode(*payload)
	if err != nil {
		logger.Errorf("failed to encode payload: '%s'", err.Error())
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, url, &buf)
	if err != nil {
		logger.Errorf("failed to create request object: '%s'", err.Error())
		return nil, fmt.Errorf("failed to create request object: '%w'", err)
	}

	req.Header.Add("Accept", accept)
	req.Header.Add("Content-Type", "application/json")

	setAuth(c, req)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return nil, fmt.Errorf("request failed: '%w'", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))

		if err == nil {
			sb.WriteString(fmt.Sprintf(": '%s'", string(body)))
		}

		errMsg := sb.String()

		logger.Error(errMsg)

		return nil, errors.New(errMsg)
	}

	logger.Infof("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	if err != nil {
		logger.Errorf("failed to read response body: '%s'", err.Error())
		return nil, fmt.Errorf("failed to read response body: '%w'", err)
	}

	var r R

	err = json.Unmarshal(body, &r)
	if err != nil {
		logger.Errorf("failed to parse response body: '%s'", err.Error())
		return nil, fmt.Errorf("failed to parse response body: '%w'", err)
	}

	return &r, nil
}

func Delete(c *Client, endpoint, accept, module string) (string, error) {
	url := c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}
	logger := l.WithFields(logrus.Fields{
		"METHOD": http.MethodDelete,
		"URL":    url,
	})

	req, err := http.NewRequestWithContext(c.ctx, http.MethodDelete, url, nil)
	if err != nil {
		logger.Errorf("failed to create request object: '%s'", err.Error())
		return "", fmt.Errorf("failed to create request object: '%w'", err)
	}

	req.Header.Add("Accept", accept)

	setAuth(c, req)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return "", fmt.Errorf("request failed: '%w'", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var sb strings.Builder

		sb.WriteString(fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))

		if err == nil {
			sb.WriteString(fmt.Sprintf(": '%s'", string(body)))
		}

		errMsg := sb.String()

		logger.Error(errMsg)

		return "", errors.New(errMsg)
	}

	logger.Infof("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	if err != nil {
		logger.Errorf("failed to read response body: '%s'", err.Error())
		return "", fmt.Errorf("failed to read response body: '%w'", err)
	}

	var resMsg struct {
		Message string `json:"message"`
	}

	err = json.Unmarshal(body, &resMsg)
	if err != nil {
		logger.Errorf("failed to parse response body: '%s'", err.Error())
		return "", fmt.Errorf("failed to parse response body: '%w'", err)
	}

	return resMsg.Message, nil
}
