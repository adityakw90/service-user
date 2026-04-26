package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

func TestAdapter_Observer_NewAuthObserver(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewAuthObserver(logger, tracer)

	if obs == nil {
		t.Fatal("NewAuthObserver() returned nil")
	}
}

func TestAdapter_Observer_AuthObserver_OnSignal_Success(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewAuthObserver(logger, tracer)
	ctx := context.Background()

	identifier := "test@example.com"
	identifierType := "email"
	fingerprint := "fp123"
	deviceName := "iPhone"

	obs.OnSignal(ctx, signal.SignalSuccess, signal.AuthSignal{
		Identifier:        identifier,
		IdentifierType:    identifierType,
		DeviceFingerprint: &fingerprint,
		DeviceName:        &deviceName,
	}, nil)

	// Verify debug log was called
	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	if entry.fields["signal"] != signal.SignalSuccess {
		t.Errorf("signal = %v, want %v", entry.fields["signal"], signal.SignalSuccess)
	}
	if entry.fields["identifier"] != "test@example.com" {
		t.Errorf("identifier = %v, want test@example.com", entry.fields["identifier"])
	}
	if entry.fields["identifier_type"] != "email" {
		t.Errorf("identifier_type = %v, want email", entry.fields["identifier_type"])
	}
}

func TestAdapter_Observer_AuthObserver_OnSignal_WithUserFields(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewAuthObserver(logger, tracer)
	ctx := context.Background()

	uid := "user-123"
	email := "user@example.com"
	username := "testuser"
	status := model.UserStatusActive
	active := true
	deleted := false

	obs.OnSignal(ctx, signal.SignalSuccess, signal.AuthSignal{
		UID:      &uid,
		Email:    &email,
		Username: &username,
		Status:   &status,
		Active:   &active,
		Deleted:  &deleted,
	}, nil)

	// Verify debug log was called
	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	if entry.fields["user.uid"] != "user-123" {
		t.Errorf("user.uid = %v, want user-123", entry.fields["user.uid"])
	}
	if entry.fields["user.email"] != "user@example.com" {
		t.Errorf("user.email = %v, want user@example.com", entry.fields["user.email"])
	}
	if entry.fields["user.active"] != true {
		t.Errorf("user.active = %v, want true", entry.fields["user.active"])
	}
}

func TestAdapter_Observer_AuthObserver_OnSignal_WithError(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewAuthObserver(logger, tracer)
	ctx := context.Background()

	identifier := "invalid@example.com"

	obs.OnSignal(ctx, signal.SignalFail, signal.AuthSignal{
		Identifier: identifier,
	}, testErr("invalid credentials"))

	// Verify error log was called
	if len(logger.errorMessages) != 1 {
		t.Fatalf("expected 1 error log, got %d", len(logger.errorMessages))
	}

	entry := logger.errorMessages[0]
	if entry.fields["signal"] != signal.SignalFail {
		t.Errorf("signal = %v, want %v", entry.fields["signal"], signal.SignalFail)
	}
}

func TestAdapter_Observer_AuthObserver_OnSignal_WithExtra(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewAuthObserver(logger, tracer)
	ctx := context.Background()

	identifier := "test@example.com"
	extra := map[string]any{"key": "value", "number": 123}

	obs.OnSignal(ctx, signal.SignalStart, signal.AuthSignal{
		Identifier: identifier,
		Extra:      &extra,
	}, nil)

	// Verify debug log was called with extra field
	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	// Extra should be JSON serialized
	if entry.fields["extra"] == nil {
		t.Error("extra field should be present in debug log")
	}
}
