package client

import (
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

func NewClient(server *url.URL) *Client {
	return &Client{
		ServerURL: server,
		HttpClient: &http.Client{
			Timeout: time.Minute,
		},
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

	var url = c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}
	logger := l.WithFields(logrus.Fields{
		"METHOD": "GET",
		"URL":    url,
	})

	req, err := http.NewRequest(http.MethodGet, url, nil)

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
func GetETag(c *Client, endpoint, accept, module string) (string, error) {

	var url = c.ServerURL.String() + endpoint

	l := logging.NewLogger()
	if c.Debug {
		l.SetDebug()
	}
	logger := l.WithFields(logrus.Fields{
		"METHOD": "GET",
		"URL":    url,
	})

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		logger.Errorf("failed to create request object, url: %s: '%s'", url, err.Error())
		return "", fmt.Errorf("failed to create request object, url: %s: '%w'", url, err)
	}

	req.Header.Add("Accept", accept)
	setAuth(c, req)

	res, err := c.HttpClient.Do(req)
	if err != nil {
		logger.Errorf("%s", err.Error())
		return "", fmt.Errorf("request failed, url: %s: '%w'", url, err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		logger.Error(fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))
		return "", fmt.Errorf("%d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	logger.Infof("%d %s", res.StatusCode, http.StatusText(res.StatusCode))

	return res.Header.Get("ETag"), nil
}
