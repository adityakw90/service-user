package event

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPPublisher_NewHTTPPublisher(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		timeout time.Duration
		logger  monitoring.Logger
		tracer  monitoring.Tracer
	}{
		{
			name:    "Valid Constructor",
			url:     "http://localhost:8080/events",
			timeout: 5 * time.Second,
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
		},
		{
			name:    "With Zero Timeout",
			url:     "http://localhost:8080/events",
			timeout: 0,
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := NewHTTPPublisher(tt.url, tt.timeout, tt.logger, tt.tracer)

			assert.NotNil(t, publisher)
			httpPublisher, ok := publisher.(*HTTPPublisher)
			require.True(t, ok, "Publisher should be *HTTPPublisher type")
			assert.Equal(t, tt.url, httpPublisher.url)
			assert.Equal(t, tt.logger, httpPublisher.logger)
			assert.Equal(t, tt.tracer, httpPublisher.tracer)
			assert.NotNil(t, httpPublisher.client)
		})
	}
}

func TestHTTPPublisher_Name(t *testing.T) {
	publisher := NewHTTPPublisher("http://localhost:8080/events", 5*time.Second, newMockLogger(), newMockTracer())

	assert.Equal(t, "HTTPPublisher", publisher.Name())
}

func TestHTTPPublisher_Close(t *testing.T) {
	publisher := NewHTTPPublisher("http://localhost:8080/events", 5*time.Second, newMockLogger(), newMockTracer())

	err := publisher.Close()

	assert.NoError(t, err)
}

type testEventData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func TestHTTPPublisher_Publish(t *testing.T) {
	tests := []struct {
		name            string
		eventType       event.EventType
		eventData       any
		responseStatus  int
		responseBody    string
		setupServer     func(t *testing.T, receivedHeaders *http.Header, receivedBody *[]byte) *httptest.Server
		wantErr         bool
		errContains     string
		validateRequest func(t *testing.T, r *http.Request)
	}{
		{
			name:           "Happy Path - 200 OK",
			eventType:      event.EventUserCreated,
			eventData:      testEventData{UserID: "123", Username: "testuser", Email: "test@example.com"},
			responseStatus: http.StatusOK,
			responseBody:   "",
			wantErr:        false,
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/cloudevents+json", r.Header.Get("Content-Type"))
				assert.Equal(t, string(event.EventUserCreated), r.Header.Get("Ce-Type"))
				assert.Equal(t, Source, r.Header.Get("Ce-Source"))
				assert.Equal(t, SpecVersion, r.Header.Get("Ce-Specversion"))
				assert.NotEmpty(t, r.Header.Get("Ce-ID"))

				var receivedCloudEventData CloudEventData
				err := json.NewDecoder(r.Body).Decode(&receivedCloudEventData)
				require.NoError(t, err)
				assert.NotEmpty(t, receivedCloudEventData.MetaData)
			},
		},
		{
			name:           "Happy Path - 202 Accepted",
			eventType:      event.EventLogin,
			eventData:      map[string]string{"user_id": "123", "ip": "192.168.1.1"},
			responseStatus: http.StatusAccepted,
			responseBody:   "",
			wantErr:        false,
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/cloudevents+json", r.Header.Get("Content-Type"))
			},
		},
		{
			name:           "4xx Client Error - 400 Bad Request",
			eventType:      event.EventDeviceCreated,
			eventData:      testEventData{UserID: "123", Username: "testuser", Email: "test@example.com"},
			responseStatus: http.StatusBadRequest,
			responseBody:   "Invalid event format",
			wantErr:        true,
			errContains:    "400",
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
			},
		},
		{
			name:           "4xx Client Error - 404 Not Found",
			eventType:      event.EventUserUpdated,
			eventData:      map[string]string{"user_id": "123"},
			responseStatus: http.StatusNotFound,
			responseBody:   "Endpoint not found",
			wantErr:        true,
			errContains:    "404",
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
			},
		},
		{
			name:           "5xx Server Error - 500 Internal Server Error",
			eventType:      event.EventUserCreated,
			eventData:      testEventData{UserID: "123", Username: "testuser", Email: "test@example.com"},
			responseStatus: http.StatusInternalServerError,
			responseBody:   "Internal server error",
			wantErr:        true,
			errContains:    "500",
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
			},
		},
		{
			name:           "5xx Server Error - 503 Service Unavailable",
			eventType:      event.EventLoginFailed,
			eventData:      map[string]string{"reason": "invalid credentials"},
			responseStatus: http.StatusServiceUnavailable,
			responseBody:   "Service temporarily unavailable",
			wantErr:        true,
			errContains:    "503",
			validateRequest: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedHeaders http.Header

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedHeaders = r.Header.Clone()

				if tt.validateRequest != nil {
					tt.validateRequest(t, r)
				} else {
					_ = json.NewDecoder(r.Body).Decode(&map[string]interface{}{})
				}

				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			logger := newMockLogger()
			tracer := newMockTracer()
			publisher := NewHTTPPublisher(server.URL, 5*time.Second, logger, tracer)

			ctx := util.SetClientName(context.Background(), "test-client")
			ctx = util.SetActor(ctx, "test-actor-id", "test-actor-type")

			err := publisher.Publish(ctx, tt.eventType, tt.eventData)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				if tt.responseBody != "" {
					assert.Contains(t, err.Error(), tt.responseBody)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NotNil(t, receivedHeaders)
		})
	}
}

func TestHTTPPublisher_Publish_ContextValues(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(ctx context.Context) context.Context
		expectedClient string
		expectedActor  struct {
			id  string
			typ string
		}
	}{
		{
			name: "With Client and Actor Context",
			setupContext: func(ctx context.Context) context.Context {
				ctx = util.SetClientName(ctx, "web-app")
				ctx = util.SetActor(ctx, "user-123", "user")
				return ctx
			},
			expectedClient: "web-app",
			expectedActor:  struct{ id, typ string }{id: "user-123", typ: "user"},
		},
		{
			name: "With No Context Values",
			setupContext: func(ctx context.Context) context.Context {
				return ctx
			},
			expectedClient: "unknown",
			expectedActor:  struct{ id, typ string }{id: "unknown", typ: "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedCloudEventData CloudEventData

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := json.NewDecoder(r.Body).Decode(&receivedCloudEventData)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			logger := newMockLogger()
			tracer := newMockTracer()
			publisher := NewHTTPPublisher(server.URL, 5*time.Second, logger, tracer)

			ctx := tt.setupContext(context.Background())

			err := publisher.Publish(ctx, event.EventUserCreated, testEventData{UserID: "123"})

			require.NoError(t, err)
			assert.Equal(t, tt.expectedClient, receivedCloudEventData.Client)
			assert.Equal(t, tt.expectedActor.id, receivedCloudEventData.ActorId)
			assert.Equal(t, tt.expectedActor.typ, receivedCloudEventData.ActorType)
		})
	}
}

func TestHTTPPublisher_Publish_TraceContextInjection(t *testing.T) {
	var traceParent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceParent = r.Header.Get("Traceparent")
		w.WriteHeader(http.StatusOK)
		log.Printf("traceParent received: `%s`", traceParent)
	}))
	defer server.Close()

	logger := newMockLogger()
	tracer := newMockTracer()
	publisher := NewHTTPPublisher(server.URL, 5*time.Second, logger, tracer)

	ctx := context.Background()
	err := publisher.Publish(ctx, event.EventUserCreated, testEventData{UserID: "123"})

	assert.NoError(t, err)
	assert.NotEmpty(t, traceParent, "Traceparent header should be injected by tracer")
}
