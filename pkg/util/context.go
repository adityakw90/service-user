package util

import (
	"context"
)

type ContextKey string

const clientNameKey ContextKey = "client-name"

func SetClientName(ctx context.Context, clientName string) context.Context {
	return context.WithValue(ctx, clientNameKey, clientName)
}

func GetClientName(ctx context.Context) string {
	if clientName, ok := ctx.Value(clientNameKey).(string); ok {
		return clientName
	}
	return "unknown"
}
