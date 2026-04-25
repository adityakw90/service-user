package util

import (
	"context"
)

type ContextKey string

const clientNameKey ContextKey = "client-name"
const actorIdKey ContextKey = "actor-id"
const actorTypeKey ContextKey = "actor-type"

func SetClientName(ctx context.Context, clientName string) context.Context {
	return context.WithValue(ctx, clientNameKey, clientName)
}

func GetClientName(ctx context.Context) string {
	if clientName, ok := ctx.Value(clientNameKey).(string); ok {
		return clientName
	}
	return "unknown"
}

func SetActor(ctx context.Context, actorId string, actorType string) context.Context {
	return context.WithValue(
		context.WithValue(ctx, actorIdKey, actorId),
		actorTypeKey, actorType,
	)
}

func GetActor(ctx context.Context) (actorId string, actorType string) {
	var ok bool
	actorId, ok = ctx.Value(actorIdKey).(string)
	if !ok {
		actorId = "unknown"
	}
	actorType, ok = ctx.Value(actorTypeKey).(string)
	if !ok {
		actorType = "unknown"
	}
	return actorId, actorType
}
