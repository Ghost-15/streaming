package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

var testActiveSessionID = "session-1"

func audioEngine(h *handler.StreamHandler, withClaims bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if withClaims {
		engine.Use(func(c *gin.Context) {
			c.Set("claims", &entity.JWTClaims{UserID: "owner", Role: entity.RoleDiffuseur})
			c.Next()
		})
	}
	engine.GET("/streams/:id/listen", h.StreamAudio)
	engine.PUT("/streams/:id/audio", h.IngestAudio)
	engine.POST("/streams/:id/push", h.PushAudio)
	return engine
}

func audioRecorder(engine *gin.Engine, method, contentType string, body []byte) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, "/streams/stream-1/audio", bytes.NewReader(body))
	if method == http.MethodGet {
		request = httptest.NewRequest(method, "/streams/stream-1/listen", nil)
	} else if method == http.MethodPost {
		request = httptest.NewRequest(method, "/streams/stream-1/push", bytes.NewReader(body))
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	request.Header.Set("X-Stream-Session-ID", testActiveSessionID)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}

func TestPushAudioValidationAndSessionReuse(t *testing.T) {
	uc := usecase.NewStreamUseCase(ownedLiveRepo(), nil)
	hub := streaming.NewHub()
	h := handler.NewStreamHandler(uc, handler.WithAudioStreaming(hub, time.Minute, time.Second, time.Second, 4, 1024, 1))

	response := audioRecorder(audioEngine(h, false), http.MethodPost, "audio/webm;codecs=opus", []byte{1})
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing claims status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(handler.NewStreamHandler(uc), true), http.MethodPost, "audio/webm;codecs=opus", []byte{1})
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing hub status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPost, "application/json", []byte{1})
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("invalid media status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPost, "audio/webm", nil)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("empty body status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPost, "audio/webm", bytes.Repeat([]byte{1}, 8))
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body status = %d", response.Code)
	}

	response = audioRecorder(audioEngine(h, true), http.MethodPost, "audio/webm;codecs=opus", []byte{1, 2})
	if response.Code != http.StatusNoContent {
		t.Fatalf("first push status = %d, want 204", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPost, "audio/webm;codecs=opus", []byte{3, 4})
	if response.Code != http.StatusNoContent {
		t.Fatalf("reused push status = %d, want 204", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPost, "audio/mpeg", []byte{5})
	if response.Code != http.StatusConflict {
		t.Fatalf("format change status = %d, want 409", response.Code)
	}
}

func TestPushAudioAuthorizesPublisherOnlyOnce(t *testing.T) {
	var lookups atomic.Int32
	repo := ownedLiveRepo()
	repo.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) {
		lookups.Add(1)
		return &entity.Stream{
			ID:              "stream-1",
			BroadcasterID:   "owner",
			Status:          entity.StreamStatusLive,
			ActiveSessionID: &testActiveSessionID,
		}, nil
	}
	hub := streaming.NewHub()
	h := handler.NewStreamHandler(
		usecase.NewStreamUseCase(repo, nil),
		handler.WithAudioStreaming(hub, time.Minute, time.Second, time.Second, 1<<20, 1024, 1),
	)
	engine := audioEngine(h, true)

	for _, chunk := range [][]byte{{1, 2}, {3, 4}, {5, 6}} {
		response := audioRecorder(engine, http.MethodPost, "audio/webm;codecs=opus", chunk)
		if response.Code != http.StatusNoContent {
			t.Fatalf("push status = %d, want 204", response.Code)
		}
	}
	if got := lookups.Load(); got != 1 {
		t.Fatalf("publisher authorization lookups = %d, want 1", got)
	}
}

func ownedLiveRepo() *mock.MockStreamRepository {
	return &mock.MockStreamRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
			return &entity.Stream{ID: "stream-1", BroadcasterID: "owner", Status: entity.StreamStatusLive, ActiveSessionID: &testActiveSessionID}, nil
		},
		IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return nil },
	}
}

func TestStreamAudioEarlyErrors(t *testing.T) {
	uc := usecase.NewStreamUseCase(ownedLiveRepo(), nil)

	response := audioRecorder(audioEngine(handler.NewStreamHandler(uc), false), http.MethodGet, "", nil)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing claims status = %d", response.Code)
	}

	response = audioRecorder(audioEngine(handler.NewStreamHandler(uc), true), http.MethodGet, "", nil)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing hub status = %d", response.Code)
	}

	hub := streaming.NewHub()
	h := handler.NewStreamHandler(uc, handler.WithAudioStreaming(hub, time.Minute, time.Second, time.Second, 1<<20, 1024, 1))
	response = audioRecorder(audioEngine(h, true), http.MethodGet, "", nil)
	if response.Code != http.StatusConflict {
		t.Fatalf("missing publisher status = %d", response.Code)
	}

	notFoundUC := usecase.NewStreamUseCase(&mock.MockStreamRepository{
		FindByIDFn: func(context.Context, string) (*entity.Stream, error) { return nil, nil },
	}, nil)
	notFoundHub := streaming.NewHub()
	if _, err := notFoundHub.OpenPublisher("stream-1", "audio/mpeg"); err != nil {
		t.Fatal(err)
	}
	notFoundHandler := handler.NewStreamHandler(notFoundUC, handler.WithAudioStreaming(notFoundHub, time.Minute, time.Second, time.Second, 1<<20, 1024, 1))
	response = audioRecorder(audioEngine(notFoundHandler, true), http.MethodGet, "", nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("join not found status = %d", response.Code)
	}
	notFoundHub.Shutdown()
}

func TestIngestAudioValidationAndConflicts(t *testing.T) {
	uc := usecase.NewStreamUseCase(ownedLiveRepo(), nil)
	hub := streaming.NewHub()
	h := handler.NewStreamHandler(uc, handler.WithAudioStreaming(hub, time.Minute, time.Second, time.Second, 4, 1024, 1))

	response := audioRecorder(audioEngine(h, false), http.MethodPut, "audio/mpeg", nil)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing claims status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(handler.NewStreamHandler(uc), true), http.MethodPut, "audio/mpeg", nil)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing hub status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPut, "application/json", nil)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("invalid media status = %d", response.Code)
	}

	if _, err := hub.OpenPublisher("stream-1", "audio/mpeg"); err != nil {
		t.Fatal(err)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPut, "audio/mpeg", nil)
	if response.Code != http.StatusConflict {
		t.Fatalf("duplicate publisher status = %d", response.Code)
	}
	hub.ClosePublisher("stream-1")

	response = audioRecorder(audioEngine(h, true), http.MethodPut, "audio/mpeg", bytes.Repeat([]byte{1}, 8))
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body status = %d, want 413", response.Code)
	}

	closedHub := streaming.NewHub()
	closedHub.Shutdown()
	closedHandler := handler.NewStreamHandler(uc, handler.WithAudioStreaming(closedHub, time.Minute, time.Second, time.Second, 1<<20, 1024, 1))
	response = audioRecorder(audioEngine(closedHandler, true), http.MethodPut, "audio/mpeg", nil)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("closed hub status = %d", response.Code)
	}
}

func TestIngestAudioRejectsNonOwner(t *testing.T) {
	repo := ownedLiveRepo()
	repo.FindByIDFn = func(context.Context, string) (*entity.Stream, error) {
		return &entity.Stream{ID: "stream-1", BroadcasterID: "someone-else", Status: entity.StreamStatusLive, ActiveSessionID: &testActiveSessionID}, nil
	}
	hub := streaming.NewHub()
	h := handler.NewStreamHandler(
		usecase.NewStreamUseCase(repo, nil),
		handler.WithAudioStreaming(hub, time.Minute, time.Second, time.Second, 1<<20, 1024, 1),
	)
	response := audioRecorder(audioEngine(h, true), http.MethodPut, "audio/mpeg", nil)
	if response.Code != http.StatusForbidden {
		t.Fatalf("non-owner status = %d", response.Code)
	}
	response = audioRecorder(audioEngine(h, true), http.MethodPost, "audio/webm", []byte{1})
	if response.Code != http.StatusForbidden {
		t.Fatalf("non-owner push status = %d", response.Code)
	}
}
