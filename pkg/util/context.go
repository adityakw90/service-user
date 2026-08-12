package util

import (
	"context"
)

type ContextKey string

const clientNameKey ContextKey = "client-name"
const actorIdKey ContextKey = "actor-id"
const actorTypeKey ContextKey = "actor-type"
const actorNameKey ContextKey = "actor-name"

func SetClientName(ctx context.Context, clientName string) context.Context {
	return context.WithValue(ctx, clientNameKey, clientName)
}

func GetClientName(ctx context.Context) string {
	if clientName, ok := ctx.Value(clientNameKey).(string); ok {
		return clientName
	}
	return "unknown"
}

func SetActor(ctx context.Context, actorId string, actorType string, actorName string) context.Context {
	return context.WithValue(
		context.WithValue(
			context.WithValue(ctx, actorIdKey, actorId),
			actorTypeKey, actorType,
		),
		actorNameKey, actorName,
	)
}

func GetActor(ctx context.Context) (actorId string, actorType string, actorName string) {
	var ok bool
	actorId, ok = ctx.Value(actorIdKey).(string)
	if !ok {
		actorId = "unknown"
	}
	actorType, ok = ctx.Value(actorTypeKey).(string)
	if !ok {
		actorType = "unknown"
	}
	actorName, ok = ctx.Value(actorNameKey).(string)
	if !ok {
		actorName = "unknown"
	}
	return actorId, actorType, actorName
}
