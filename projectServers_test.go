package solus

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectsService_ServersCreate(t *testing.T) {
	data := ProjectServersCreateRequest{
		Name:             "name",
		PlanID:           1,
		LocationID:       2,
		OsImageVersionID: 3,
		SSHKeys:          []int{4, 5},
		UserData:         "user data",
	}

	s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects/42/servers", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assertRequestBody(t, r, data)

		writeResponse(t, w, http.StatusCreated, fakeVirtualServer)
	})
	defer s.Close()

	actual, err := createTestClient(t, s.URL).Projects.ServersCreate(context.Background(), 42, data)
	require.NoError(t, err)
	require.Equal(t, fakeVirtualServer, actual)
}

func TestProjectsService_ServersListAll(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		page := int32(0)

		s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			p := atomic.LoadInt32(&page)

			assert.Equal(t, "/projects/1/servers", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			if page == 0 {
				assert.Equal(t, "", r.URL.Query().Get("page"))
			} else {
				assert.Equal(t, strconv.Itoa(int(p)), r.URL.Query().Get("page"))
			}

			if p == 2 {
				writeJSON(t, w, http.StatusOK, ProjectServersResponse{
					Data: []VirtualServer{
						{
							ID: int(p),
						},
					},
					paginatedResponse: paginatedResponse{
						Links: ResponseLinks{
							Next: r.URL.String(),
						},
						Meta: ResponseMeta{
							CurrentPage: int(p),
							LastPage:    2,
						},
					},
				})
				return
			}
			atomic.AddInt32(&page, 1)

			q := r.URL.Query()
			q.Set("page", strconv.Itoa(int(p)+1))
			r.URL.RawQuery = q.Encode()

			writeJSON(t, w, http.StatusOK, ProjectServersResponse{
				paginatedResponse: paginatedResponse{
					Links: ResponseLinks{
						Next: r.URL.String(),
					},
					Meta: ResponseMeta{
						CurrentPage: 1,
						LastPage:    2,
					},
				},
				Data: []VirtualServer{{ID: int(p)}},
			})
		})
		defer s.Close()

		c := createTestClient(t, s.URL)

		actual, err := c.Projects.ServersListAll(context.Background(), 1)
		require.NoError(t, err)

		require.Equal(t, []VirtualServer{
			{ID: 0},
			{ID: 1},
			{ID: 2},
		}, actual)
	})

	t.Run("negative", func(t *testing.T) {
		s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/projects/1/servers", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "", r.URL.Query().Get("page"))
			w.WriteHeader(http.StatusBadRequest)
		})
		defer s.Close()

		_, err := createTestClient(t, s.URL).Projects.ServersListAll(context.Background(), 1)
		assert.EqualError(t, err, "HTTP GET projects/1/servers returns 400 status code")
	})
}

func TestProjectsService_Servers(t *testing.T) {
	expected := ProjectServersResponse{
		Data: []VirtualServer{
			fakeVirtualServer,
		},
	}

	s := startTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects/42/servers", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		writeJSON(t, w, http.StatusOK, expected)
	})
	defer s.Close()

	actual, err := createTestClient(t, s.URL).Projects.Servers(context.Background(), 42)
	require.NoError(t, err)
	actual.service = nil
	require.Equal(t, expected, actual)
}
