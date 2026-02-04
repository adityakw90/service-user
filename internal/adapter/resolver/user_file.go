package resolver

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/adityakw90/go-monitoring"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	portMonitoring "github.com/adityakw90/service-user/internal/core/port/monitoring"
	portResolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type userFileResolver struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             portMonitoring.Logger
	tracer             monitoring.Tracer
}

type userFileIdentity struct {
	id  int64
	uid string
}

func NewUserFileResolver(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger portMonitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.UserFileResolver {
	return &userFileResolver{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (r *userFileResolver) fetchIDFromDB(ctx context.Context, uid string) (*userFileIdentity, error) {
	var iden userFileIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM user_file WHERE uid=$1`, uid,
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
		return nil, domainerrors.ErrFileNotFound
	}

	return &iden, nil
}

func (r *userFileResolver) fetchUIDFromDB(ctx context.Context, id int64) (*userFileIdentity, error) {
	var iden userFileIdentity

	rows, err := r.db.Query(ctx,
		`SELECT id, uid FROM user_file WHERE id=$1`, id,
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
		return nil, domainerrors.ErrFileNotFound
	}

	return &iden, nil
}

func (r *userFileResolver) IDsByUIDs(ctx context.Context, userFileUIDs []string) (map[string]int64, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "userFileResolver.IDsByUIDs")
	defer resvSpan.End()

	result, err := mapperID(
		newCtx,
		r.logger,
		r.redisClient,
		userFileUIDs,
		func(res string) int64 {
			d, _ := strconv.ParseInt(res, 10, 64)
			return d
		},
		func(uid string) string {
			return r.redisPrefix + ":" + uid + ":id"
		},
		func(uid string) (*userFileIdentity, error) {
			return r.fetchIDFromDB(newCtx, uid)
		},
		func(userFile *userFileIdentity) int64 {
			return userFile.id
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrFileNotFound) {
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
		attribute.StringSlice("userFileUID", userFileUIDs),
	))

	return result, nil
}

func (r *userFileResolver) UIDsByIDs(ctx context.Context, userFileIDs []int64) (map[int64]string, error) {
	newCtx, resvSpan := r.tracer.StartSpan(ctx, "userFileResolver.UIDsByIDs")
	defer resvSpan.End()

	result, err := mapperUID(
		newCtx,
		r.logger,
		r.redisClient,
		userFileIDs,
		func(res string) string { return res },
		func(id int64) string {
			return r.redisPrefix + ":id:" + strconv.FormatInt(id, 10) + ":uid"
		},
		func(id int64) (*userFileIdentity, error) {
			return r.fetchUIDFromDB(newCtx, id)
		},
		func(userFile *userFileIdentity) string {
			return userFile.uid
		},
		r.redisCacheDuration,
	)
	if err != nil {
		if errors.Is(err, domainerrors.ErrFileNotFound) {
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
		attribute.Int64Slice("userFileID", userFileIDs),
	))

	return result, nil
}
