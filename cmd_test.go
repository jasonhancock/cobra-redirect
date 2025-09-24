package redirect

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jasonhancock/go-logger"
	"github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tests := []struct {
		desc     string
		dest     string
		input    string
		expected string
	}{
		{
			"normal",
			"https://127.0.0.1",
			"/foo",
			"https://127.0.0.1/foo",
		},
		{
			"dest with path",
			"https://127.0.0.1/bar",
			"/foo",
			"https://127.0.0.1/bar/foo",
		},
		{
			"dest with port",
			"https://127.0.0.1:443/bar",
			"/foo",
			"https://127.0.0.1:443/bar/foo",
		},
		{
			"with params",
			"https://127.0.0.1",
			"/foo?bar=baz",
			"https://127.0.0.1/foo?bar=baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			serverAddr := generic.NewRandomPort(t)
			opts := options{
				HTTPAddr: serverAddr,
				DestAddr: tt.dest,
				HTTPCode: http.StatusFound,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go run(ctx, logger.Silence(), opts)
			time.Sleep(500 * time.Millisecond)

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					// Return http.ErrUseLastResponse to prevent following redirects
					return http.ErrUseLastResponse
				},
			}
			req, err := http.NewRequest(http.MethodGet, "http://"+serverAddr+tt.input, nil)
			require.NoError(t, err)
			resp, err := client.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())

			require.Equal(t, http.StatusFound, resp.StatusCode)
			require.Equal(t, tt.expected, resp.Header.Get("Location"))
		})
	}
}
