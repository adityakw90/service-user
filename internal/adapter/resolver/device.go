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

type deviceResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

func NewDeviceResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.DeviceResolver {
	return &deviceResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *deviceResolver) fetchIDFromDB(ctx context.Context, uid string) (*identity, error) {
	var iden identity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM device WHERE uid=$1`, uid,
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

	return &iden, nil
}

func (r *deviceResolver) IDsByUIDs(ctx context.Context, deviceUIDs []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "deviceResolver.IDsByUIDs")
	defer resvSpan.End()

	result, err := mapperID(
		newCtx,
		r.logger,
		r.redisClient,
		deviceUIDs,
		func(res string) int64 {
			d, _ := strconv.ParseInt(res, 10, 64)
			return d
		},
		func(uid string) string {
			return r.redisPrefix + ":" + uid + ":id"
		},
		r.fetchIDFromDB,
		func(device *identity) int64 {
			return device.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrDeviceNotFound) {
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
		attribute.StringSlice("deviceUID", deviceUIDs),
	))

	return result, nil
}

func (r *deviceResolver) fetchUIDFromDB(ctx context.Context, id int64) (*identity, error) {
	var iden identity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM device WHERE id=$1`, id,
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
		return nil, domainerrors.ErrDeviceNotFound
	}

	return &iden, nil
}

func (r *deviceResolver) UIDsByIDs(ctx context.Context, deviceIDs []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "deviceResolver.UIDsByIDs")
	defer resvSpan.End()

	result, err := mapperUID(
		newCtx,
		r.logger,
		r.redisClient,
		deviceIDs,
		func(res string) string { return res },
		func(id int64) string {
			return r.redisPrefix + ":id:" + strconv.FormatInt(id, 10) + ":uid"
		},
		r.fetchUIDFromDB,
		func(device *identity) string {
			return device.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrDeviceNotFound) {
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
		attribute.Int64Slice("deviceID", deviceIDs),
	))

	return result, nil
}

// Invalidate clears cached entries for the specified UIDs/IDs.
func (r *deviceResolver) Invalidate(ctx context.Context, opts ...params.InvalidateOpt) error {
	// TODO: Implement cache invalidation
	return nil
}
