package resolver

import (
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	portResolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	"github.com/redis/go-redis/v9"
)

type resolverProvider struct {
	db                 PostgrePool
	redisClient        *redis.Client
	redisPrefix        string
	redisCacheDuration time.Duration
	logger             monitoring.Logger
	tracer             monitoring.Tracer
}

func NewResolverProvider(
	db PostgrePool,
	redisClient *redis.Client,
	redisPrefix string,
	redisCacheDuration time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portResolver.ResolverProvider {
	return &resolverProvider{
		db:                 db,
		redisClient:        redisClient,
		redisPrefix:        redisPrefix,
		redisCacheDuration: redisCacheDuration,
		logger:             logger,
		tracer:             tracer,
	}
}

func (p *resolverProvider) User() portResolver.UserResolver {
	return NewUserResolver(
		p.db,
		p.redisClient,
		p.redisPrefix+":user",
		p.redisCacheDuration,
		p.logger,
		p.tracer,
	)
}

func (p *resolverProvider) Device() portResolver.DeviceResolver {
	return NewDeviceResolver(
		p.db,
		p.redisClient,
		p.redisPrefix+":device",
		p.redisCacheDuration,
		p.logger,
		p.tracer,
	)
}
