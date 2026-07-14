package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func liveStream() *entity.Stream {
	return &entity.Stream{ID: "s1", Title: "Live", BroadcasterID: "bc-1", Status: entity.StreamStatusLive}
}

func TestStreamUseCase_ListActive(t *testing.T) {
	t.Run("success returns streams", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			ListActiveFn: func(_ context.Context) ([]entity.Stream, error) {
				return []entity.Stream{{ID: "s1"}, {ID: "s2"}}, nil
			},
		}
		uc := usecase.NewStreamUseCase(repo, nil)
		streams, err := uc.ListActive(context.Background())
		if err != nil {
			t.Fatalf("ListActive() err = %v", err)
		}
		if len(streams) != 2 {
			t.Errorf("ListActive() len = %d, want 2", len(streams))
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			ListActiveFn: func(_ context.Context) ([]entity.Stream, error) {
				return nil, errors.New("db down")
			},
		}
		uc := usecase.NewStreamUseCase(repo, nil)
		if _, err := uc.ListActive(context.Background()); err == nil {
			t.Error("ListActive() expected error")
		}
	})
}

func TestStreamUseCase_Start(t *testing.T) {
	tests := []struct {
		name      string
		bcaster   string
		title     string
		repoSetup func(*mock.MockStreamRepository)
		wantErr   error
	}{
		{
			name:    "success",
			bcaster: "bc-1",
			title:   "My Live",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.CreateFn = func(_ context.Context, s *entity.Stream) error { s.ID = "s1"; return nil }
			},
		},
		{name: "empty broadcaster", bcaster: "", title: "x", wantErr: usecase.ErrStreamInvalid},
		{name: "empty title", bcaster: "bc-1", title: "  ", wantErr: usecase.ErrStreamInvalid},
		{
			name:    "repo error",
			bcaster: "bc-1",
			title:   "My Live",
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
			uc := usecase.NewStreamUseCase(repo, nil)
			s, err := uc.Start(context.Background(), tt.bcaster, tt.title)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Start() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err == nil && s == nil {
				t.Error("Start() returned nil stream on success")
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
			bcaster: "bc-1",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil }
				r.UpdateStatusFn = func(_ context.Context, _ string, _ entity.StreamStatus) error { return nil }
			},
		},
		{
			name:    "not found",
			bcaster: "bc-1",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return nil, nil }
			},
			wantErr: usecase.ErrStreamNotFound,
		},
		{
			name:    "forbidden",
			bcaster: "other",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil }
			},
			wantErr: usecase.ErrStreamForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			tt.repoSetup(repo)
			uc := usecase.NewStreamUseCase(repo, nil)
			err := uc.End(context.Background(), "s1", tt.bcaster)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("End() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestStreamUseCase_Join(t *testing.T) {
	t.Run("success records history", func(t *testing.T) {
		recorded := false
		repo := &mock.MockStreamRepository{
			FindByIDFn:           func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil },
			IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return nil },
		}
		history := &mock.MockListenHistoryRepository{
			RecordFn: func(_ context.Context, _ *entity.ListenHistory) error { recorded = true; return nil },
		}
		uc := usecase.NewStreamUseCase(repo, history)
		if err := uc.Join(context.Background(), "s1", "u1"); err != nil {
			t.Fatalf("Join() err = %v", err)
		}
		if !recorded {
			t.Error("Join() did not record listen history")
		}
	})

	t.Run("empty ids", func(t *testing.T) {
		uc := usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil)
		if err := uc.Join(context.Background(), "", "u1"); !errors.Is(err, usecase.ErrStreamInvalid) {
			t.Fatalf("Join() err = %v, want ErrStreamInvalid", err)
		}
	})

	t.Run("stream not live", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) { return nil, nil },
		}
		uc := usecase.NewStreamUseCase(repo, nil)
		if err := uc.Join(context.Background(), "s1", "u1"); !errors.Is(err, usecase.ErrStreamNotFound) {
			t.Fatalf("Join() err = %v, want ErrStreamNotFound", err)
		}
	})
}

func TestStreamUseCase_Leave(t *testing.T) {
	repo := &mock.MockStreamRepository{
		IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return nil },
	}
	uc := usecase.NewStreamUseCase(repo, nil)
	if err := uc.Leave(context.Background(), "s1", "u1"); err != nil {
		t.Fatalf("Leave() err = %v", err)
	}
	if err := uc.Leave(context.Background(), "", "u1"); !errors.Is(err, usecase.ErrStreamInvalid) {
		t.Fatalf("Leave() empty err = %v, want ErrStreamInvalid", err)
	}
}
