package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

const (
	bcasterID    = "broadcaster-1"
	otherBcaster = "broadcaster-2"
	testStreamID = "stream-1"
	demoURL      = "https://example.com/audio.mp3"
)

func liveStream() *entity.Stream {
	return &entity.Stream{
		ID:            testStreamID,
		Title:         "Live show",
		BroadcasterID: bcasterID,
		StreamURL:     demoURL,
		Status:        entity.StreamStatusLive,
	}
}

func TestStreamUseCase_Start(t *testing.T) {
	tests := []struct {
		name      string
		bcaster   string
		title     string
		url       string
		repoSetup func(*mock.MockStreamRepository)
		wantErr   error
	}{
		{
			name:    "success",
			bcaster: bcasterID,
			title:   "Live show",
			url:     demoURL,
			repoSetup: func(r *mock.MockStreamRepository) {
				r.CreateFn = func(_ context.Context, s *entity.Stream) error {
					s.ID = testStreamID
					return nil
				}
			},
		},
		{name: "empty broadcaster", bcaster: "", title: "x", url: demoURL, wantErr: usecase.ErrStreamInvalid},
		{name: "empty title", bcaster: bcasterID, title: "  ", url: demoURL, wantErr: usecase.ErrStreamInvalid},
		{
			name:    "repo error",
			bcaster: bcasterID,
			title:   "Live show",
			url:     demoURL,
			repoSetup: func(r *mock.MockStreamRepository) {
				r.CreateFn = func(_ context.Context, _ *entity.Stream) error { return errors.New("db") }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewStreamUseCase(repo)
			s, err := uc.Start(context.Background(), tt.bcaster, tt.title, tt.url)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Start() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err == nil && s != nil && s.StreamURL != demoURL {
				t.Errorf("Start() url = %q, want %q", s.StreamURL, demoURL)
			}
		})
	}
}

func TestStreamUseCase_End(t *testing.T) {
	tests := []struct {
		name      string
		bcaster   string
		repoSetup func(*mock.MockStreamRepository)
		wantErr   error
	}{
		{
			name:    "success",
			bcaster: bcasterID,
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil }
				r.UpdateStatusFn = func(_ context.Context, _ string, _ entity.StreamStatus) error { return nil }
			},
		},
		{
			name:    "not found",
			bcaster: bcasterID,
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return nil, nil }
			},
			wantErr: usecase.ErrStreamNotFound,
		},
		{
			name:    "forbidden",
			bcaster: otherBcaster,
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil }
			},
			wantErr: usecase.ErrStreamForbidden,
		},
		{
			name:    "already ended",
			bcaster: bcasterID,
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) {
					s := liveStream()
					s.Status = entity.StreamStatusEnded
					return s, nil
				}
			},
			wantErr: usecase.ErrStreamEnded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			tt.repoSetup(repo)
			uc := usecase.NewStreamUseCase(repo)
			err := uc.End(context.Background(), testStreamID, tt.bcaster)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("End() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestStreamUseCase_ListActive(t *testing.T) {
	repo := &mock.MockStreamRepository{
		ListActiveFn: func(_ context.Context) ([]entity.Stream, error) {
			return []entity.Stream{*liveStream()}, nil
		},
	}
	uc := usecase.NewStreamUseCase(repo)
	streams, err := uc.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive() err = %v", err)
	}
	if len(streams) != 1 {
		t.Errorf("ListActive() len = %d, want 1", len(streams))
	}
}

func TestStreamUseCase_Join(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		repoSetup func(*mock.MockStreamRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: "listener-1",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil }
				r.IncrementListenersFn = func(_ context.Context, _ string, _ int) error { return nil }
			},
		},
		{name: "empty user", userID: "", repoSetup: func(_ *mock.MockStreamRepository) {}, wantErr: usecase.ErrStreamInvalid},
		{
			name:   "not found",
			userID: "listener-1",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return nil, nil }
			},
			wantErr: usecase.ErrStreamNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			tt.repoSetup(repo)
			uc := usecase.NewStreamUseCase(repo)
			s, err := uc.Join(context.Background(), testStreamID, tt.userID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Join() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err == nil && s.ListenerCount != 1 {
				t.Errorf("Join() listenerCount = %d, want 1", s.ListenerCount)
			}
		})
	}
}

func TestStreamUseCase_Leave(t *testing.T) {
	repo := &mock.MockStreamRepository{
		IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return nil },
	}
	uc := usecase.NewStreamUseCase(repo)
	if err := uc.Leave(context.Background(), testStreamID, "listener-1"); err != nil {
		t.Fatalf("Leave() err = %v", err)
	}
	if err := uc.Leave(context.Background(), testStreamID, ""); !errors.Is(err, usecase.ErrStreamInvalid) {
		t.Fatalf("Leave() empty user err = %v, want ErrStreamInvalid", err)
	}
}
