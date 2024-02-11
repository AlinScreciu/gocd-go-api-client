package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AlinScreciu/gocd-go-api-client/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Transport interface {
	RoundTrip(req *http.Request) (*http.Response, error)
}
type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Simulate a network error
	return nil, errors.New("network error")
}

func TestClient_SetAuth(t *testing.T) {
	t.Parallel()
	type args struct {
		user     string
		password string
		token    string
	}
	tests := []struct {
		name     string
		client   *Client
		authType AuthType
		args     args
	}{
		{
			name:   "NewClientShouldHaveAuthTypeNone",
			client: NewClient(context.TODO(), &url.URL{}),
		},
		{
			name:     "Should set basic auth",
			client:   NewClient(context.TODO(), &url.URL{}),
			authType: Basic,
			args: args{
				user:     "user",
				password: "pass",
			},
		},
		{
			name:     "Should set accessToken auth",
			client:   NewClient(context.TODO(), &url.URL{}),
			authType: AccessToken,
			args: args{
				token: "token",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			switch tt.authType {
			case None:
				assert.Equal(t, None, tt.client.auth)
				assert.Empty(t, tt.client.user)
				assert.Empty(t, tt.client.password)
				assert.Empty(t, tt.client.token)
			case Basic:
				tt.client.SetBasicAuth(tt.args.user, tt.args.password)

				assert.Equal(t, Basic, tt.client.auth)
				assert.Equal(t, tt.args.user, tt.client.user)
				assert.Equal(t, tt.args.password, tt.client.password)
				assert.Empty(t, tt.client.token)

			case AccessToken:
				tt.client.SetAccessToken(tt.args.token)

				assert.Equal(t, AccessToken, tt.client.auth)
				assert.Equal(t, tt.args.token, tt.client.token)
				assert.Empty(t, tt.client.user)
				assert.Empty(t, tt.client.password)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	type args struct {
		server *url.URL
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "NewClient",
			args: args{server: &url.URL{Host: "fake.com"}},
			want: &Client{
				ServerURL: &url.URL{Host: "fake.com"},
				HttpClient: &http.Client{
					Timeout: time.Minute,
				},
				ctx: context.TODO(),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewClient(context.TODO(), tt.args.server)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_setAuth(t *testing.T) {
	t.Parallel()
	type args struct {
		auth     AuthType
		token    string
		password string
		user     string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestBasicAuth",
			args: args{
				auth:     Basic,
				user:     "test",
				password: "1234",
			},
		},
		{
			name: "TestBearerTokenAuth",
			args: args{
				auth:  AccessToken,
				token: "1234",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := NewClient(context.TODO(), &url.URL{})
			switch tt.args.auth {
			case Basic:
				client.SetBasicAuth(tt.args.user, tt.args.password)
			case AccessToken:
				client.SetAccessToken(tt.args.token)
			}
			req, err := http.NewRequestWithContext(client.ctx, http.MethodGet, "https://fake.com", nil)
			require.NoError(t, err)

			setAuth(client, req)

			authHeader := req.Header.Get("Authorization")

			switch tt.args.auth {
			case Basic:
				want := "Basic " + base64.StdEncoding.EncodeToString([]byte(tt.args.user+":"+tt.args.password))
				assert.Equal(t, want, authHeader)
			case AccessToken:
				assert.Equal(t, "Bearer "+tt.args.token, authHeader)
			}
		})
	}
}

type Version struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Doc struct {
			Href string `json:"href"`
		} `json:"doc"`
	} `json:"_links"`
	Version     string `json:"version"`
	BuildNumber string `json:"build_number"`
	GitSha      string `json:"git_sha"`
	FullVersion string `json:"full_version"`
	CommitURL   string `json:"commit_url"`
}

func TestGet(t *testing.T) {
	t.Parallel()
	type args struct {
		ts *httptest.Server
	}
	tests := []struct {
		name    string
		args    args
		want    *Version
		wantErr bool
	}{
		{
			name: "TestValid",
			want: &Version{
				Links: struct {
					Self struct {
						Href string `json:"href"`
					} `json:"self"`
					Doc struct {
						Href string `json:"href"`
					} `json:"doc"`
				}{
					Self: struct {
						Href string `json:"href"`
					}{
						Href: "https://build.go.cd/go/api/version",
					},
					Doc: struct {
						Href string `json:"href"`
					}{
						Href: "https://api.gocd.org/#version",
					},
				},
				Version:     "16.6.0",
				BuildNumber: "3348",
				GitSha:      "a7a5717cbd60c30006314fb8dd529796c93adaf0",
				FullVersion: "16.6.0 (3348-a7a5717cbd60c30006314fb8dd529796c93adaf0)",
				CommitURL:   "https://github.com/gocd/gocd/commits/a7a5717cbd60c30006314fb8dd529796c93adaf0",
			},
			args: args{
				ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`{
						"_links": {
						  "self": {
							"href": "https://build.go.cd/go/api/version"
						  },
						  "doc": {
							"href": "https://api.gocd.org/#version"
						  }
						},
						"version": "16.6.0",
						"build_number": "3348",
						"git_sha": "a7a5717cbd60c30006314fb8dd529796c93adaf0",
						"full_version": "16.6.0 (3348-a7a5717cbd60c30006314fb8dd529796c93adaf0)",
						"commit_url": "https://github.com/gocd/gocd/commits/a7a5717cbd60c30006314fb8dd529796c93adaf0"
					  }`))
					if err != nil {
						t.Errorf("failed to write body: '%s'", err.Error())
					}
				})),
			},
		},
		{
			name:    "TestErrorResponse",
			wantErr: true,
			args: args{
				ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				})),
			},
		},
		{
			name:    "TestInvalidResponseBody",
			wantErr: true,
			args: args{
				ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(`invalid json response`))
					if err != nil {
						t.Errorf("failed to write body: '%s'", err.Error())
					}
				})),
			},
		},
		{
			name:    "TestNoResponseBody",
			wantErr: true,
			args: args{
				ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.WriteHeader(http.StatusOK)
				})),
			},
		},
		{
			name:    "TestRequestFailure",
			wantErr: true,
			args: args{
				ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Server doesn't respond, causing request failure
				})),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer tt.args.ts.Close()
			url, _ := url.Parse(tt.args.ts.URL)
			got, err := Get[Version](NewClient(context.TODO(), url), "/", constants.AcceptV1, "test")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequestFailureWithTimeout(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a long-running process that exceeds the client timeout
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	c := NewClient(context.TODO(), url)
	c.HttpClient.Timeout = 1 * time.Second // Set the client timeout to 1 second

	// Test the Get function
	_, err := Get[Version](c, "/", constants.AcceptV1, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Client.Timeout")

	// Test the GetETag function
	_, _, err = GetWithETag[Version](c, "/", constants.AcceptV1, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Client.Timeout")
}

func TestNon2xxSuccess(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a 3xx response which should be treated as an error by the client
		w.WriteHeader(http.StatusMovedPermanently)
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	c := NewClient(context.TODO(), url)

	// Test the Get function
	_, err := Get[Version](c, "/", constants.AcceptV1, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "301 Moved Permanently")

	// Test the GetETag function
	_, etag, err := GetWithETag[Version](c, "/", constants.AcceptV1, "test")
	assert.Empty(t, etag)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "301 Moved Permanently")
}

func TestClientErrorMalformedURL(t *testing.T) {
	t.Parallel()
	_, err := Get[Version](NewClient(context.TODO(), &url.URL{}), "/", constants.AcceptV1, "test")
	require.Error(t, err)
	// The error message should indicate the URL is invalid
	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}

func TestGetETag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		handler   func(w http.ResponseWriter, r *http.Request)
		wantETag  string
		wantErr   bool
		transport Transport
	}{
		{
			name: "TestValidETag",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("ETag", `"123456"`)
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{
					"_links": {
					  "self": {
						"href": "https://build.go.cd/go/api/version"
					  },
					  "doc": {
						"href": "https://api.gocd.org/#version"
					  }
					},
					"version": "16.6.0",
					"build_number": "3348",
					"git_sha": "a7a5717cbd60c30006314fb8dd529796c93adaf0",
					"full_version": "16.6.0 (3348-a7a5717cbd60c30006314fb8dd529796c93adaf0)",
					"commit_url": "https://github.com/gocd/gocd/commits/a7a5717cbd60c30006314fb8dd529796c93adaf0"
				  }`))
				if err != nil {
					t.Errorf("failed to write body: '%s'", err.Error())
				}
			},
			transport: &http.Transport{},
			wantETag:  `"123456"`,
			wantErr:   false,
		},
		{
			name: "TestNoETag",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantETag:  "",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "TestErrorResponse",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantETag:  "",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "TestNetworkError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantETag:  "",
			wantErr:   true,
			transport: &mockTransport{},
		},
		{
			name: "TestBodyError",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Length", "100")
				w.WriteHeader(http.StatusOK)
			},
			wantETag:  "",
			wantErr:   true,
			transport: &http.Transport{},
		},
		// More test cases can be added here
		{
			name: "TestNoETag",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{
					"_links": {
					  "self": {
						"href": "https://build.go.cd/go/api/version"
					  },
					  "doc": {
						"href": "https://api.gocd.org/#version"
					  }
					},
					"version": "16.6.0",
					"build_number": "3348",
					"git_sha": "a7a5717cbd60c30006314fb8dd529796c93adaf0",
					"full_version": "16.6.0 (3348-a7a5717cbd60c30006314fb8dd529796c93adaf0)",
					"commit_url": "https://github.com/gocd/gocd/commits/a7a5717cbd60c30006314fb8dd529796c93adaf0"
				  }`))
				if err != nil {
					t.Errorf("failed to write body: '%s'", err.Error())
				}
			},
			wantETag:  "",
			wantErr:   true,
			transport: &http.Transport{},
		},
		// More test cases can be added here
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer ts.Close()

			url, _ := url.Parse(ts.URL)
			c := NewClient(context.TODO(), url)
			c.HttpClient.Transport = tt.transport

			_, gotETag, err := GetWithETag[Version](c, "/", constants.AcceptV1, "test")

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Expected no error but got one")
			}

			assert.Equal(t, tt.wantETag, gotETag, "ETag did not match expected value")
		})
	}
}

func TestPut(t *testing.T) {
	t.Parallel()

	type Payload struct {
		Message string  `json:"message"`
		Number  float64 `json:"number"`
	}
	samplePayload := &Payload{
		Message: "test",
		Number:  10,
	}

	tests := []struct {
		name         string
		setupHandler func(t *testing.T) http.HandlerFunc
		payload      *Payload
		eTag         string
		endpoint     string
		want         *Payload
		wantErr      bool
		ctx          context.Context
		transport    Transport
	}{
		{
			name: "Successful PUT with valid ETag and payload",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, "test-etag", r.Header.Get("If-Match"))

					var v Payload
					err := json.NewDecoder(r.Body).Decode(&v)
					require.NoError(t, err)
					require.Equal(t, samplePayload, &v)

					w.Header().Set("Content-Type", constants.AcceptV1)
					err = json.NewEncoder(w).Encode(samplePayload)
					require.NoError(t, err)
				}
			},
			payload:   samplePayload,
			eTag:      "test-etag",
			endpoint:  "/test-endpoint",
			want:      samplePayload,
			wantErr:   false,
			ctx:       context.Background(),
			transport: &http.Transport{},
		},
		{
			name: "PUT with mismatched ETag",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusPreconditionFailed)
				}
			},
			payload:   samplePayload,
			eTag:      "mismatched-etag",
			endpoint:  "/test-endpoint",
			wantErr:   true,
			transport: &http.Transport{},

			ctx: context.Background(),
		},
		{
			name: "PUT with server error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-endpoint",
			transport: &http.Transport{},

			wantErr: true,
			ctx:     context.Background(),
		},
		{
			name: "PUT with false content length",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.WriteHeader(http.StatusOK)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-endpoint",
			transport: &http.Transport{},

			wantErr: true,
			ctx:     context.Background(),
		},
		{
			name: "PUT with malformed JSON response",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", constants.AcceptV1)
					_, _ = w.Write([]byte(`malformed json`))
				}
			},
			transport: &http.Transport{},
			payload:   &Payload{},
			eTag:      "valid-etag",
			endpoint:  "/test-endpoint",
			wantErr:   true,
			ctx:       context.Background(),
		},
		{
			name: "PUT with malformed JSON response",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", constants.AcceptV1)
					_, _ = w.Write([]byte(`malformed json`))
				}
			},
			payload:   &Payload{Number: math.NaN()},
			transport: &http.Transport{},
			eTag:      "valid-etag",
			endpoint:  "/test-endpoint",
			wantErr:   true,
			ctx:       context.Background(),
		},
		{
			name: "PUT with network error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", constants.AcceptV1)
					_, _ = w.Write([]byte(`malformed json`))
				}
			},
			transport: &mockTransport{},
			payload:   samplePayload,
			eTag:      "valid-etag",
			endpoint:  "/test-endpoint",
			wantErr:   true,
			ctx:       context.Background(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tt.setupHandler(t))
			defer server.Close()

			client := NewClient(context.Background(), &url.URL{Scheme: "http", Host: server.Listener.Addr().String()})
			client.HttpClient.Transport = tt.transport

			got, err := Put[Payload, Payload](client, tt.payload, tt.eTag, tt.endpoint, constants.AcceptV1, "test-module")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPost(t *testing.T) {
	t.Parallel()

	// Define your payload and response types
	type Payload struct {
		Message string  `json:"message"`
		Number  float64 `json:"number"`
	}
	type Response struct {
		Result string `json:"result"`
	}

	samplePayload := &Payload{
		Message: "Hello, world!",
	}
	sampleResponse := &Response{
		Result: "Success",
	}

	tests := []struct {
		name         string
		setupHandler func(t *testing.T) http.HandlerFunc
		payload      *Payload
		endpoint     string
		want         *Response
		wantErr      bool
		transport    Transport
	}{
		{
			name: "Successful POST",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					var p Payload
					err := json.NewDecoder(r.Body).Decode(&p)
					require.NoError(t, err)
					require.Equal(t, samplePayload, &p)

					w.Header().Set("Content-Type", "application/json")
					err = json.NewEncoder(w).Encode(sampleResponse)
					require.NoError(t, err)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-post",
			want:      sampleResponse,
			wantErr:   false,
			transport: &http.Transport{},
		},
		{
			name: "POST with server error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-post",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "POST with network error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-post",
			wantErr:   true,
			transport: &mockTransport{},
		},
		{
			name: "POST with body error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.WriteHeader(http.StatusOK)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-post",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "POST with network error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-post",
			wantErr:   true,
			transport: &mockTransport{},
		},
		{
			name: "POST with malformed JSON response",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`malformed json`))
				}
			},
			payload:   samplePayload,
			endpoint:  "/test-post",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "POST with malformed payload",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`malformed json`))
				}
			},
			payload: func() *Payload {
				return &Payload{
					Number: math.NaN(),
				}
			}(),
			endpoint:  "/test-post",
			wantErr:   true,
			transport: &http.Transport{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tt.setupHandler(t))
			defer server.Close()

			client := NewClient(context.Background(), &url.URL{Scheme: "http", Host: server.Listener.Addr().String()})
			client.HttpClient.Transport = tt.transport

			got, err := Post[Payload, Response](client, tt.payload, tt.endpoint, "application/json", "test-module")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupHandler func(t *testing.T) http.HandlerFunc
		endpoint     string
		want         string
		wantErr      bool
		transport    Transport
	}{
		{
			name: "Successful DELETE",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// Simulate a response for successful deletion
					_, err := w.Write([]byte(`{"message": "Delete operation successful"}`))
					require.NoError(t, err)
				}
			},
			endpoint:  "/test-delete",
			want:      `Delete operation successful`,
			wantErr:   false,
			transport: &http.Transport{},
		},
		{
			name: "Successful DELETE but response isn't valid",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					// Simulate a response for successful deletion
					_, err := w.Write([]byte(`Delete operation successful`))
					require.NoError(t, err)
				}
			},
			endpoint:  "/test-delete",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "DELETE with server error",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			endpoint:  "/test-delete",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "DELETE with not found",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}
			},
			endpoint:  "/test-delete",
			wantErr:   true,
			transport: &http.Transport{},
		},
		{
			name: "DELETE with fail on read body",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.WriteHeader(http.StatusOK)
				}
			},
			transport: &http.Transport{},

			endpoint: "/test-delete",
			wantErr:  true,
		},
		{
			name: "DELETE with timeout",
			setupHandler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
				}
			},
			transport: &mockTransport{},

			endpoint: "/test-delete",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tt.setupHandler(t))
			defer server.Close()

			client := NewClient(context.Background(), &url.URL{Scheme: "http", Host: server.Listener.Addr().String()})
			client.HttpClient.Transport = tt.transport

			got, err := Delete(client, tt.endpoint, "application/json", "test-module")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
