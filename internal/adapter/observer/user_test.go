package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

func TestAdapter_Observer_NewUserObserver(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewUserObserver(logger, tracer)

	if obs == nil {
		t.Fatal("NewUserObserver() returned nil")
	}
}

func TestAdapter_Observer_UserObserver_OnSignal_Success(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewUserObserver(logger, tracer)
	ctx := context.Background()

	uid := "user-123"
	username := "testuser"
	status := model.UserStatusActive
	active := true

	obs.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &uid,
		Username:  &username,
		Status:    &status,
		Active:    &active,
		Operation: "get",
	}, nil)

	// Verify debug log was called
	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	if entry.fields["signal"] != signal.SignalSuccess {
		t.Errorf("signal = %v, want %v", entry.fields["signal"], signal.SignalSuccess)
	}
	if entry.fields["operation"] != "get" {
		t.Errorf("operation = %v, want get", entry.fields["operation"])
	}
}

func TestAdapter_Observer_UserObserver_OnSignal_WithError(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewUserObserver(logger, tracer)
	ctx := context.Background()

	uid := "user-456"

	obs.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
		UID:       &uid,
		Operation: "get",
	}, testErr("user not found"))

	// Verify error log was called
	if len(logger.errorMessages) != 1 {
		t.Fatalf("expected 1 error log, got %d", len(logger.errorMessages))
	}
}

type testErr string

func (e testErr) Error() string {
	return string(e)
}
