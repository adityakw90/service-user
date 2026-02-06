package resolver

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/adityakw90/go-monitoring"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	portResolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type userResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

type userIdentity struct {
	id  int64
	uid string
}

func NewUserResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.UserResolver {
	return &userResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *userResolver) fetchIDFromDB(ctx context.Context, uid string) (*userIdentity, error) {
	var iden userIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "user" WHERE uid=$1`, uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&iden.id, &iden.uid)
		if err != nil {
			return nil, err
		}
	}

	if iden.id == 0 {
		return nil, domainerrors.ErrUserNotFound
	}

	return &iden, nil
}

func (r *userResolver) fetchUIDFromDB(ctx context.Context, id int64) (*userIdentity, error) {
	var iden userIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM "user" WHERE id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&iden.id, &iden.uid)
		if err != nil {
			return nil, err
		}
	}

	if iden.id == 0 {
		return nil, domainerrors.ErrUserNotFound
	}

	return &iden, nil
}

func (r *userResolver) IDsByUIDs(ctx context.Context, userUIDs []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "userResolver.IDsByUIDs")
	defer resvSpan.End()

	result, err := mapperID(
		newCtx,
		r.logger,
		r.redisClient,
		userUIDs,
		func(res string) int64 {
			d, _ := strconv.ParseInt(res, 10, 64)
			return d
		},
		func(uid string) string {
			return r.redisPrefix + ":" + uid + ":id"
		},
		func(uid string) (*userIdentity, error) {
			return r.fetchIDFromDB(newCtx, uid)
		},
		func(user *userIdentity) int64 {
			return user.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			r.logger.Debug("Failed", map[string]interface{}{
				"error.message": err.Error(),
			})
		} else {
			r.logger.Error("error", map[string]interface{}{
				"error.message": err.Error(),
			})
		}
		resvSpan.AddEvent("Error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	resvSpan.AddEvent("success", trace.WithAttributes(
		attribute.StringSlice("userUID", userUIDs),
	))

	return result, nil
}

func (r *userResolver) UIDsByIDs(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "userResolver.UIDsByIDs")
	defer resvSpan.End()

	result, err := mapperUID(
		newCtx,
		r.logger,
		r.redisClient,
		userIDs,
		func(res string) string { return res },
		func(id int64) string {
			return r.redisPrefix + ":id:" + strconv.FormatInt(id, 10) + ":uid"
		},
		func(id int64) (*userIdentity, error) {
			return r.fetchUIDFromDB(newCtx, id)
		},
		func(user *userIdentity) string {
			return user.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			r.logger.Debug("Failed", map[string]interface{}{
				"error.message": err.Error(),
			})
		} else {
			r.logger.Error("error", map[string]interface{}{
				"error.message": err.Error(),
			})
		}
		resvSpan.AddEvent("Error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	resvSpan.AddEvent("success", trace.WithAttributes(
		attribute.Int64Slice("userID", userIDs),
	))

	return result, nil
}
