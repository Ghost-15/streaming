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
	sessionID := "session-1"
	return &entity.Stream{ID: "s1", Title: "Live", BroadcasterID: "bc-1", Status: entity.StreamStatusLive, ActiveSessionID: &sessionID}
}

func TestStreamUseCase_ListOwned(t *testing.T) {
	repo := &mock.MockStreamRepository{
		ListByBroadcasterFn: func(_ context.Context, broadcasterID string) ([]entity.Stream, error) {
			if broadcasterID != "bc-1" {
				t.Fatalf("broadcasterID = %q", broadcasterID)
			}
			return []entity.Stream{{ID: "s1"}}, nil
		},
	}
	streams, err := usecase.NewStreamUseCase(repo, nil).ListOwned(context.Background(), "bc-1")
	if err != nil || len(streams) != 1 {
		t.Fatalf("ListOwned streams=%v err=%v", streams, err)
	}
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

func TestStreamUseCase_Restart(t *testing.T) {
	var activatedSession string
	repo := &mock.MockStreamRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
			return &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusEnded}, nil
		},
		ActivateFn: func(_ context.Context, id, sessionID string) error {
			if id != "s1" || sessionID == "" {
				t.Fatalf("Activate(%q, %q)", id, sessionID)
			}
			activatedSession = sessionID
			return nil
		},
	}
	stream, err := usecase.NewStreamUseCase(repo, nil).Restart(context.Background(), "s1", "bc-1")
	if err != nil || stream == nil || !stream.IsLive() || stream.ActiveSessionID == nil {
		t.Fatalf("Restart stream=%v err=%v", stream, err)
	}
	if *stream.ActiveSessionID != activatedSession {
		t.Fatalf("active session = %q, want %q", *stream.ActiveSessionID, activatedSession)
	}
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

func TestStreamUseCase_DeleteOwned(t *testing.T) {
	deleted := false
	repo := &mock.MockStreamRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
			return &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusEnded}, nil
		},
		DeleteFn: func(_ context.Context, id string) error {
			deleted = id == "s1"
			return nil
		},
	}
	if err := usecase.NewStreamUseCase(repo, nil).Delete(context.Background(), "s1", "bc-1"); err != nil || !deleted {
		t.Fatalf("Delete deleted=%v err=%v", deleted, err)
	}
}

func TestStreamUseCase_CanBroadcastRequiresCurrentSession(t *testing.T) {
	uc := usecase.NewStreamUseCase(&mock.MockStreamRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil },
	}, nil)
	if err := uc.CanBroadcast(context.Background(), "s1", "bc-1", "session-1"); err != nil {
		t.Fatalf("current session rejected: %v", err)
	}
	if err := uc.CanBroadcast(context.Background(), "s1", "bc-1", "stale"); !errors.Is(err, usecase.ErrStreamSessionExpired) {
		t.Fatalf("stale session err = %v", err)
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
				r.DeactivateFn = func(_ context.Context, _, _ string) (bool, error) { return true, nil }
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
			err := uc.End(context.Background(), "s1", tt.bcaster, "session-1")
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("End() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestStreamUseCase_End_Branches(t *testing.T) {
	session := "session-1"
	liveOwned := func(sess string) *entity.Stream {
		s := &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusLive}
		if sess != "" {
			s.ActiveSessionID = &sess
		}
		return s
	}

	t.Run("empty session", func(t *testing.T) {
		uc := usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil)
		if err := uc.End(context.Background(), "s1", "bc-1", ""); !errors.Is(err, usecase.ErrStreamInvalid) {
			t.Fatalf("End() err = %v, want ErrStreamInvalid", err)
		}
	})

	t.Run("already ended returns nil", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
				return &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusEnded}, nil
			},
		}
		if err := usecase.NewStreamUseCase(repo, nil).End(context.Background(), "s1", "bc-1", session); err != nil {
			t.Fatalf("End() on ended stream err = %v, want nil", err)
		}
	})

	t.Run("session mismatch", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) { return liveOwned("other"), nil },
		}
		if err := usecase.NewStreamUseCase(repo, nil).End(context.Background(), "s1", "bc-1", session); !errors.Is(err, usecase.ErrStreamSessionExpired) {
			t.Fatalf("End() mismatch err = %v, want ErrStreamSessionExpired", err)
		}
	})

	t.Run("deactivate reports superseded", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn:   func(_ context.Context, _ string) (*entity.Stream, error) { return liveOwned(session), nil },
			DeactivateFn: func(_ context.Context, _, _ string) (bool, error) { return false, nil },
		}
		if err := usecase.NewStreamUseCase(repo, nil).End(context.Background(), "s1", "bc-1", session); !errors.Is(err, usecase.ErrStreamSessionExpired) {
			t.Fatalf("End() superseded err = %v, want ErrStreamSessionExpired", err)
		}
	})

	t.Run("deactivate repo error", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn:   func(_ context.Context, _ string) (*entity.Stream, error) { return liveOwned(session), nil },
			DeactivateFn: func(_ context.Context, _, _ string) (bool, error) { return false, errors.New("db down") },
		}
		if err := usecase.NewStreamUseCase(repo, nil).End(context.Background(), "s1", "bc-1", session); err == nil {
			t.Fatal("End() expected repo error")
		}
	})
}

func TestStreamUseCase_Delete_Branches(t *testing.T) {
	t.Run("live stream decrements active count", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil },
			DeleteFn:   func(_ context.Context, _ string) error { return nil },
		}
		if err := usecase.NewStreamUseCase(repo, nil).Delete(context.Background(), "s1", "bc-1"); err != nil {
			t.Fatalf("Delete() live err = %v", err)
		}
	})

	t.Run("repo delete error", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
				return &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusEnded}, nil
			},
			DeleteFn: func(_ context.Context, _ string) error { return errors.New("db") },
		}
		if err := usecase.NewStreamUseCase(repo, nil).Delete(context.Background(), "s1", "bc-1"); err == nil {
			t.Fatal("Delete() expected repo error")
		}
	})

	t.Run("not owned", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil },
		}
		if err := usecase.NewStreamUseCase(repo, nil).Delete(context.Background(), "s1", "intruder"); !errors.Is(err, usecase.ErrStreamForbidden) {
			t.Fatalf("Delete() not owned err = %v, want ErrStreamForbidden", err)
		}
	})
}

func TestStreamUseCase_Join(t *testing.T) {
	t.Run("success records history", func(t *testing.T) {
		recorded := false
		repo := &mock.MockStreamRepository{
			FindByIDFn:           func(_ context.Context, _ string) (*entity.Stream, error) { return liveStream(), nil },
			IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return nil },
		}
		history := &mock.MockListenHistoryRepository{
			RecordFn: func(_ context.Context, entry *entity.ListenHistory) error {
				recorded = true
				if entry.Event != entity.ListenEventJoin {
					t.Fatalf("history event = %q, want join", entry.Event)
				}
				return nil
			},
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
	t.Run("success records history", func(t *testing.T) {
		recorded := false
		repo := &mock.MockStreamRepository{
			IncrementListenersFn: func(_ context.Context, id string, delta int) error {
				if id != "s1" || delta != -1 {
					t.Fatalf("IncrementListeners(%q, %d), want s1, -1", id, delta)
				}
				return nil
			},
		}
		history := &mock.MockListenHistoryRepository{
			RecordFn: func(_ context.Context, entry *entity.ListenHistory) error {
				recorded = true
				if entry.Event != entity.ListenEventLeave {
					t.Fatalf("history event = %q, want leave", entry.Event)
				}
				return nil
			},
		}
		uc := usecase.NewStreamUseCase(repo, history)
		if err := uc.Leave(context.Background(), "s1", "u1"); err != nil {
			t.Fatalf("Leave() err = %v", err)
		}
		if !recorded {
			t.Error("Leave() did not record listen history")
		}
	})

	t.Run("empty ids", func(t *testing.T) {
		uc := usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil)
		if err := uc.Leave(context.Background(), "", "u1"); !errors.Is(err, usecase.ErrStreamInvalid) {
			t.Fatalf("Leave() empty err = %v, want ErrStreamInvalid", err)
		}
	})
}
