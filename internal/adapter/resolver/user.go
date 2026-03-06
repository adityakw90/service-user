package resolver

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/adityakw90/go-monitoring"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/param"
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

func (r *userResolver) fetchIDFromDB(ctx context.Context, uid string) (*identity, error) {
	var iden identity

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

func (r *userResolver) fetchUIDFromDB(ctx context.Context, id int64) (*identity, error) {
	var iden identity

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
		r.fetchIDFromDB,
		func(user *identity) int64 {
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
		r.fetchUIDFromDB,
		func(user *identity) string {
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

// Invalidate clears cached entries for the specified UIDs/IDs.
func (r *userResolver) Invalidate(ctx context.Context, opts ...param.InvalidateOpt) error {
	// TODO: Implement cache invalidation
	return nil
}
