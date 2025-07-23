package probes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewProbesMuxer(t *testing.T) {
	t.Run("should return muxer with healthz and readyz endpoints", func(t *testing.T) {
		log := zap.NewNop().Sugar()

		muxer := newProbesMuxer(":1234", log)
		require.NotNil(t, muxer)
	})
}

func TestHealthz(t *testing.T) {
	t.Run("should return 200", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()
		healthz(w, r)
		require.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}

func readyzHandleSuccess(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readyzHandleFail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}
func readyzHandleServerError(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestReadyz(t *testing.T) {
	t.Run("should return 200 when reverse proxy is reachable", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		testServer := httptest.NewServer(http.HandlerFunc(readyzHandleSuccess))
		r := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		getReadyz(testServer.URL, log)(w, r)
		require.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("should return 200 status code when service returns non-5XX", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		testServer := httptest.NewServer(http.HandlerFunc(readyzHandleFail))
		r := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		getReadyz(testServer.URL, log)(w, r)
		require.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("should return status code when service returns 5XX", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		testServer := httptest.NewServer(http.HandlerFunc(readyzHandleServerError))
		r := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		getReadyz(testServer.URL, log)(w, r)
		require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	})

	t.Run("should return 503 when reverse proxy is unreachable", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		r := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		getReadyz(":0", log)(w, r)
		require.Equal(t, http.StatusServiceUnavailable, w.Result().StatusCode)
	})
}

func TestNew(t *testing.T) {
	t.Run("should return server with probes", func(t *testing.T) {
		log := zap.NewNop().Sugar()
		server := New(":1234", ":5678", log)
		require.NotNil(t, server)
	})
}
