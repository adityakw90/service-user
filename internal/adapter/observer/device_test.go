package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

func TestAdapter_Observer_NewDeviceObserver(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewDeviceObserver(logger, tracer)

	if obs == nil {
		t.Fatal("NewDeviceObserver() returned nil")
	}
}

func TestAdapter_Observer_DeviceObserver_OnSignal_Success(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewDeviceObserver(logger, tracer)
	ctx := context.Background()

	uid := "device-123"
	deviceName := "iPhone 15"

	obs.OnSignal(ctx, signal.SignalSuccess, signal.DeviceSignal{
		UID:        &uid,
		DeviceName: &deviceName,
		Operation:  "get",
	}, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	if entry.fields["operation"] != "get" {
		t.Errorf("operation = %v, want get", entry.fields["operation"])
	}
}
