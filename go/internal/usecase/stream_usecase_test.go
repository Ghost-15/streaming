package usecase_test

import (
	"context"
	"testing"

	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func TestStreamUseCase_ListActive(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "not implemented yet",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			uc := usecase.NewStreamUseCase(repo)

			_, err := uc.ListActive(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ListActive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
