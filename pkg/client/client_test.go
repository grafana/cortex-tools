package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildURL(t *testing.T) {
	tc := []struct {
		name      string
		path      string
		method    string
		url       string
		resultURL string
	}{
		{
			name:      "builds the correct URL with a trailing slash",
			path:      "/api/v1/rules",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/",
			resultURL: "http://cortexurl.com/api/v1/rules",
		},
		{
			name:      "builds the correct URL without a trailing slash",
			path:      "/api/v1/rules",
			method:    http.MethodPost,
			url:       "http://cortexurl.com",
			resultURL: "http://cortexurl.com/api/v1/rules",
		},
		{
			name:      "builds the correct URL when the base url has a path",
			path:      "/api/v1/rules",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/apathto",
			resultURL: "http://cortexurl.com/apathto/api/v1/rules",
		},
		{
			name:      "builds the correct URL when the base url has a path with trailing slash",
			path:      "/api/v1/rules",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/apathto/",
			resultURL: "http://cortexurl.com/apathto/api/v1/rules",
		},
		{
			name:      "builds the correct URL with a trailing slash and the target path contains special characters",
			path:      "/api/v1/rules/%20%2Fspace%F0%9F%8D%BB",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/",
			resultURL: "http://cortexurl.com/api/v1/rules/%20%2Fspace%F0%9F%8D%BB",
		},
		{
			name:      "builds the correct URL without a trailing slash and the target path contains special characters",
			path:      "/api/v1/rules/%20%2Fspace%F0%9F%8D%BB",
			method:    http.MethodPost,
			url:       "http://cortexurl.com",
			resultURL: "http://cortexurl.com/api/v1/rules/%20%2Fspace%F0%9F%8D%BB",
		},
		{
			name:      "builds the correct URL when the base url has a path and the target path contains special characters",
			path:      "/api/v1/rules/%20%2Fspace%F0%9F%8D%BB",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/apathto",
			resultURL: "http://cortexurl.com/apathto/api/v1/rules/%20%2Fspace%F0%9F%8D%BB",
		},
		{
			name:      "builds the correct URL when the base url has a path and the target path starts with a escaped slash",
			path:      "/api/v1/rules/%2F-first-char-slash",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/apathto",
			resultURL: "http://cortexurl.com/apathto/api/v1/rules/%2F-first-char-slash",
		},
		{
			name:      "builds the correct URL when the base url has a path and the target path ends with a escaped slash",
			path:      "/api/v1/rules/last-char-slash%2F",
			method:    http.MethodPost,
			url:       "http://cortexurl.com/apathto",
			resultURL: "http://cortexurl.com/apathto/api/v1/rules/last-char-slash%2F",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.Parse(tt.url)
			require.NoError(t, err)

			req, err := buildRequest(context.Background(), tt.path, tt.method, *url, []byte{})
			require.NoError(t, err)
			require.Equal(t, tt.resultURL, req.URL.String())
		})
	}

}

func TestDoRequest(t *testing.T) {
	requestCh := make(chan *http.Request, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCh <- r
		fmt.Fprintln(w, "hello")
	}))
	defer ts.Close()

	for _, tc := range []struct {
		name         string
		user         string
		key          string
		id           string
		authToken    string
		extraHeaders map[string]string
		expectedErr  string
		validate     func(t *testing.T, req *http.Request)
	}{
		{
			name: "basic headers only, no extra headers",
			id:   "my-tenant-id",
			validate: func(t *testing.T, req *http.Request) {
				require.Equal(t, "my-tenant-id", req.Header.Get("X-Scope-OrgID"))
				// Verify no extra headers were added
				require.Empty(t, req.Header.Get("key1"))
			},
		},
		{
			name: "extraHeaders are added",
			id:   "my-tenant-id",
			extraHeaders: map[string]string{
				"key1":          "value1",
				"key2":          "value2",
				"X-Scope-OrgID": "first-tenant-id",
			},
			validate: func(t *testing.T, req *http.Request) {
				require.Equal(t, "value1", req.Header.Get("key1"))
				require.Equal(t, "value2", req.Header.Get("key2"))
				require.Equal(t, []string{"first-tenant-id", "my-tenant-id"}, req.Header.Values("X-Scope-OrgID"))
			},
		},
		{
			name:      "auth token with extra headers",
			id:        "my-tenant-id",
			authToken: "my-auth-token",
			extraHeaders: map[string]string{
				"Custom-Header": "custom-value",
			},
			validate: func(t *testing.T, req *http.Request) {
				require.Equal(t, "Bearer my-auth-token", req.Header.Get("Authorization"))
				require.Equal(t, "my-tenant-id", req.Header.Get("X-Scope-OrgID"))
				require.Equal(t, "custom-value", req.Header.Get("Custom-Header"))
			},
		},
		{
			name:      "authorization header in extra headers is ignored when auth token is set",
			id:        "my-tenant-id",
			authToken: "my-auth-token",
			extraHeaders: map[string]string{
				"Authorization": "Bearer should-be-ignored",
			},
			validate: func(t *testing.T, req *http.Request) {
				// The Authorization header from extraHeaders should be overwritten
				require.Equal(t, "Bearer my-auth-token", req.Header.Get("Authorization"))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := New(Config{
				Address:      ts.URL,
				User:         tc.user,
				Key:          tc.key,
				AuthToken:    tc.authToken,
				ID:           tc.id,
				ExtraHeaders: tc.extraHeaders,
			})
			require.NoError(t, err)

			res, err := client.doRequest(ctx, "/test", http.MethodGet, nil)

			if tc.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, res.StatusCode)
				req := <-requestCh
				tc.validate(t, req)
			}
		})
	}
}
