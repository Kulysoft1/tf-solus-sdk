// Autogenerated file. Do not edit!

package solus

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRolesResponse_Next(t *testing.T) {
	page := int32(1)

	s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		p := atomic.LoadInt32(&page)

		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/roles", r.URL.Path)
		require.Equal(t, strconv.Itoa(int(p)), r.URL.Query().Get("page"))

		if p == 3 {
			writeJSON(t, w, http.StatusOK, RolesResponse{Data: []Role{{Id: int(p)}}})
			return
		}
		atomic.AddInt32(&page, 1)

		q := r.URL.Query()
		q.Set("page", strconv.Itoa(int(p)+1))
		r.URL.RawQuery = q.Encode()

		writeJSON(t, w, http.StatusOK, RolesResponse{
			paginatedResponse: paginatedResponse{
				Links: ResponseLinks{
					Next: r.URL.String(),
				},
			},
			Data: []Role{{Id: int(p)}},
		})
	})
	defer s.Close()

	u, err := url.Parse(s.URL)
	require.NoError(t, err)

	c, err := NewClient(u, authenticator{})
	require.NoError(t, err)

	resp := RolesResponse{
		paginatedResponse: paginatedResponse{
			Links: ResponseLinks{
				Next: fmt.Sprintf("%s/roles?page=1", s.URL),
			},
			service: &service{c},
		},
	}

	i := 1
	for resp.Next(context.Background()) {
		require.Equal(t, []Role{{Id: i}}, resp.Data)
		i++
	}
	require.NoError(t, resp.err)
	require.Equal(t, 4, i, "Expects to get 3 entity, but got less")
}
