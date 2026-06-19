package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/srvsurya/system-monitor/internal/models"
)

// success test

type FakeStatsStore struct{}

func (f FakeStatsStore) GetLatestMetric() (
	models.SystemMetric,
	error,
) {
	return models.SystemMetric{}, nil
}

func TestGetCurrentStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.GET("/stats",
		GetCurrentStats(FakeStatsStore{}))

	req := httptest.NewRequest(
		http.MethodGet,
		"/stats",
		nil,
	)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

// error test

type ErrorStatsStore struct{}

func (e ErrorStatsStore) GetLatestMetric() (
	models.SystemMetric,
	error,
) {
	return models.SystemMetric{},
		errors.New("db failed")
}

func TestGetCurrentStats_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.GET("/stats",
		GetCurrentStats(ErrorStatsStore{}))

	req := httptest.NewRequest(
		http.MethodGet,
		"/stats",
		nil,
	)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", w.Code)
	}
}
