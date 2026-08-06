package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

func TestAdapter_Observer_NewUserFileObserver(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewUserFileObserver(logger, tracer)

	if obs == nil {
		t.Fatal("NewUserFileObserver() returned nil")
	}
}

func TestAdapter_Observer_UserFileObserver_OnSignal_Success(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewUserFileObserver(logger, tracer)
	ctx := context.Background()

	uid := "file-123"
	fileName := "profile.jpg"
	fileSize := int64(1024)

	obs.OnSignal(ctx, signal.SignalSuccess, signal.UserFileSignal{
		UID:       &uid,
		FileName:  &fileName,
		FileSize:  &fileSize,
		Operation: "get",
	}, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}
}
