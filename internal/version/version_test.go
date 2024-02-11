package version

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	type args struct {
		ts *httptest.Server
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Version
		wantErr bool
	}{
		{
			name: "TestValid",
			want: &types.Version{
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

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer tt.args.ts.Close()
			url, _ := url.Parse(tt.args.ts.URL)
			got, err := GetVersion(client.NewClient(context.TODO(), url))

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
	c := client.NewClient(context.TODO(), url)
	c.HttpClient.Timeout = 1 * time.Second // Set the client timeout to 1 second

	_, err := GetVersion(c)
	require.Error(t, err)
	// The error should be related to the client timeout
	assert.Contains(t, err.Error(), "Client.Timeout")
}

func TestNon2xxSuccess(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a 3xx response which should be treated as an error by the client
		w.WriteHeader(http.StatusMovedPermanently)
		_, err := w.Write([]byte("moved"))
		if err != nil {
			t.Errorf("failed to write body: '%s'", err.Error())
		}
	}))
	defer ts.Close()

	url, _ := url.Parse(ts.URL)
	_, err := GetVersion(client.NewClient(context.TODO(), url))
	require.Error(t, err)
	// The error should contain the 3xx status code
	assert.Contains(t, err.Error(), "301 Moved Permanently")
}

func TestClientErrorMalformedURL(t *testing.T) {
	t.Parallel()
	_, err := GetVersion(client.NewClient(context.TODO(), &url.URL{})) // Pass an empty url.URL object
	require.Error(t, err)
	// The error message should indicate the URL is invalid
	assert.Contains(t, err.Error(), "unsupported protocol scheme")
}
