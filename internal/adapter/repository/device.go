package repository

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	portRepo "github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type deviceRepository struct {
	db                 *gorm.DB
	redis              *redis.Client
	tracer             monitoring.Tracer
	logger             monitoring.Logger
	columnOrderMap     map[string]bool
	columnOrderDefault string
	columnSortMap      map[string]bool
	columnSortDefault  string
	redisPrefix        string
	redisMapDuration   time.Duration
}

func NewDeviceRepository(
	db *gorm.DB,
	redisClient *redis.Client,
	monitoringInstance *monitoring.Monitoring,
	redisPrefix string,
	redisMapDuration time.Duration,
) portRepo.DeviceRepository {
	columnOrderMap := map[string]bool{
		"id":         true,
		"uid":        true,
		"device_id":  true,
		"created_at": true,
		"updated_at": true,
	}
	columnSortMap := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	return &deviceRepository{
		db:                 db,
		redis:              redisClient,
		tracer:             monitoringInstance.Tracer,
		logger:             monitoringInstance.Logger,
		columnOrderMap:     columnOrderMap,
		columnOrderDefault: "id",
		columnSortMap:      columnSortMap,
		columnSortDefault:  "asc",
		redisPrefix:        redisPrefix + ":DEVICE",
		redisMapDuration:   redisMapDuration,
	}
}

func (r *deviceRepository) Create(ctx context.Context, device *model.Device) (*model.Device, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Insert the device into the database
		if err := tx.Create(device).Error; err != nil {
			return err // rollback
		}
		return nil // commit
	})

	if err != nil {
		return nil, err
	}

	return device, nil
}

func (r *deviceRepository) Update(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the device into the database
		return tx.Save(device).Error
	})
}

func (r *deviceRepository) Delete(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete the device from the database
		return tx.Delete(device).Error
	})
}

func (r *deviceRepository) GetByID(ctx context.Context, id int64) (*model.Device, error) {
	// Create a child span
	_, repoSpan := r.tracer.StartSpan(ctx, "DeviceRepository.GetByID")
	defer repoSpan.End()

	// Add attributes to the span
	repoSpan.SetAttributes(
		attribute.Int("arg.id", int(id)),
	)

	// avoid stale data in pgbouncer transaction mode
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback() // Rollback since it's a read-only transaction

	var device model.Device

	if err := tx.First(&device, "id = ?", id).Error; err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrDeviceNotFound
		}
		return nil, err
	}

	repoSpan.AddEvent("success", trace.WithAttributes(
		attribute.Int("device.id", int(device.ID)),
	))

	return &device, nil
}

func (r *deviceRepository) GetByUID(ctx context.Context, uid string) (*model.Device, error) {
	// Create a child span
	_, repoSpan := r.tracer.StartSpan(ctx, "DeviceRepository.GetByUID")
	defer repoSpan.End()

	// Add attributes to the span
	repoSpan.SetAttributes(
		attribute.String("arg.uid", uid),
	)

	// avoid stale data in pgbouncer transaction mode
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback() // Rollback since it's a read-only transaction

	var device model.Device

	if err := tx.First(&device, "uid = ?", uid).Error; err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrDeviceNotFound
		}
		return nil, err
	}

	repoSpan.AddEvent("success", trace.WithAttributes(
		attribute.String("device.uid", uid),
	))

	return &device, nil
}

func (r *deviceRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.Device, error) {
	// Create a child span
	_, repoSpan := r.tracer.StartSpan(ctx, "DeviceRepository.GetByFingerprint")
	defer repoSpan.End()

	// Add attributes to the span
	repoSpan.SetAttributes(
		attribute.String("arg.fingerprint", fingerprint),
	)

	// avoid stale data in pgbouncer transaction mode
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback() // Rollback since it's a read-only transaction

	var device model.Device

	if err := tx.First(&device, "fingerprint = ?", fingerprint).Error; err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrDeviceNotFound
		}
		return nil, err
	}

	repoSpan.AddEvent("success", trace.WithAttributes(
		attribute.String("device.fingerprint", fingerprint),
	))

	return &device, nil
}

func (r *deviceRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	// Create a child span
	_, repoSpan := r.tracer.StartSpan(ctx, "DeviceRepository.List")
	defer repoSpan.End()

	var devices []model.Device
	var total int64
	var err error
	page := 1
	limit := 0
	totalPages := 1

	// avoid stale data in pgbouncer transaction mode
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback() // Rollback since it's a read-only transaction

	// Build the query based on filter conditions
	queryBuilder := tx.Model(&model.Device{})

	if len(filter.Ids) > 0 {
		queryBuilder = queryBuilder.Where("id IN ?", filter.Ids)
	}
	if len(filter.Uids) > 0 {
		queryBuilder = queryBuilder.Where("uid IN ?", filter.Uids)
	}
	if filter.DeviceName != nil {
		queryBuilder = queryBuilder.Where("device_name = ?", filter.DeviceName)
	}

	// Get total count before applying pagination
	err = queryBuilder.Count(&total).Error
	if err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	if pagination.OrderBy != nil {
		orderBy := *pagination.OrderBy
		sort := r.columnSortDefault
		if !r.columnOrderMap[orderBy] {
			orderBy = r.columnOrderDefault
		}
		if pagination.Sort != nil {
			sort = *pagination.Sort
		}
		if !r.columnSortMap[sort] {
			sort = r.columnSortDefault
		}
		queryBuilder = queryBuilder.Order(orderBy + " " + sort)
	}
	if pagination.Limit != nil {
		limit = *pagination.Limit
		if pagination.Page != nil {
			page = *pagination.Page
		}
		if page == 0 {
			page = 1
		}
		if limit == 0 {
			limit = 10
		}
		queryBuilder = queryBuilder.Offset((page - 1) * limit).Limit(limit)
		if total > 0 {
			totalPages = int(math.Ceil(float64(total) / float64(limit)))
		}
	}

	// Apply ordering, pagination, and fetch results
	err = queryBuilder.Find(&devices).Error

	if err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	repoSpan.AddEvent("success", trace.WithAttributes(
		attribute.Int("result.page", page),
		attribute.Int("result.limit", limit),
		attribute.Int64("result.total", total),
		attribute.Int("result.pages", totalPages),
	))

	// return devices, total, nil
	return &model.Devices{
		Items: devices,
		Meta: model.Meta{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: totalPages,
		},
	}, nil
}

func (r *deviceRepository) ListByUserID(ctx context.Context, userId int64, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	// Create a child span
	_, repoSpan := r.tracer.StartSpan(ctx, "DeviceRepository.List")
	defer repoSpan.End()

	var devices []model.Device
	var total int64
	var err error
	page := 1
	limit := 0
	totalPages := 1

	// avoid stale data in pgbouncer transaction mode
	tx := r.db.WithContext(ctx).Begin()
	defer tx.Rollback() // Rollback since it's a read-only transaction

	queryBuilder := tx.Model(&model.Device{}).
		Joins(`JOIN "user_device" ON "user_device"."device_id" = "device"."id"`).
		Where(`"user_device"."user_id" = ?`, userId)

	if len(filter.Ids) > 0 {
		queryBuilder = queryBuilder.Where("id IN ?", filter.Ids)
	}
	if len(filter.Uids) > 0 {
		queryBuilder = queryBuilder.Where("uid IN ?", filter.Uids)
	}
	if filter.DeviceName != nil {
		queryBuilder = queryBuilder.Where("device_name = ?", filter.DeviceName)
	}

	// Get total count before applying pagination
	err = queryBuilder.Count(&total).Error
	if err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	if pagination.OrderBy != nil {
		orderBy := *pagination.OrderBy
		sort := r.columnSortDefault
		if !r.columnOrderMap[orderBy] {
			orderBy = r.columnOrderDefault
		}
		if pagination.Sort != nil {
			sort = *pagination.Sort
		}
		if !r.columnSortMap[sort] {
			sort = r.columnSortDefault
		}
		queryBuilder = queryBuilder.Order(orderBy + " " + sort)
	}
	if pagination.Limit != nil {
		limit = *pagination.Limit
		if pagination.Page != nil {
			page = *pagination.Page
		}
		if page == 0 {
			page = 1
		}
		if limit == 0 {
			limit = 10
		}
		queryBuilder = queryBuilder.Offset((page - 1) * limit).Limit(limit)
		if total > 0 {
			totalPages = int(math.Ceil(float64(total) / float64(limit)))
		}
	}

	// Apply ordering, pagination, and fetch results
	err = queryBuilder.Find(&devices).Error

	if err != nil {
		repoSpan.AddEvent("failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	repoSpan.AddEvent("success", trace.WithAttributes(
		attribute.Int("result.page", page),
		attribute.Int("result.limit", limit),
		attribute.Int64("result.total", total),
		attribute.Int("result.pages", totalPages),
	))

	// return devices, total, nil
	return &model.Devices{
		Items: devices,
		Meta: model.Meta{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: totalPages,
		},
	}, nil
}

func (r *deviceRepository) MapIDsByUIDs(ctx context.Context, deviceUIDs []string) (map[string]int64, error) {
	// Create a child span
	ctx, repoSpan := r.tracer.StartSpan(ctx, "deviceRepository.ListIDsByUIDs")
	defer repoSpan.End()

	// get logger with spancontext
	logger := r.logger.WithSpanContext(repoSpan.SpanContext())

	result, err := mapperID(
		ctx,
		r.tracer,
		r.logger,
		r.redis,
		deviceUIDs,
		func(res string) int64 {
			d, _ := strconv.ParseInt(res, 10, 64)
			return d
		},
		func(uid string) string {
			return r.redisPrefix + ":" + uid + ":id"
		},
		func(uid string) (*model.Device, error) {
			var device model.Device

			// avoid stale data in pgbouncer transaction mode
			tx := r.db.WithContext(ctx).Begin()
			defer tx.Rollback() // Rollback since it's a read-only transaction

			if err := tx.Select("id, uid").Where("uid = ?", uid).First(&device).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, domainErrors.ErrDeviceNotFound
				}
				return nil, err
			}
			return &device, nil
		},
		func(device *model.Device) int64 {
			return device.ID
		},
		r.redisMapDuration,
	)
	if err != nil {
		if errors.Is(err, domainErrors.ErrDeviceNotFound) {
			logger.Debug("Failed", map[string]interface{}{
				"error.message": err.Error(),
			})
		} else {
			logger.Error("error", map[string]interface{}{
				"error.message": err.Error(),
			})
		}
		repoSpan.AddEvent("Error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	repoSpan.AddEvent("success", trace.WithAttributes(
		attribute.StringSlice("deviceUID", deviceUIDs),
	))

	return result, nil
}

func (r *deviceRepository) MapUIDsByIDs(ctx context.Context, deviceIDs []int64) (map[int64]string, error) {
	// Create a child span
	ctx, repoSpan := r.tracer.StartSpan(ctx, "deviceRepository.ListIDsByUIDs")
	defer repoSpan.End()

	// get logger with spancontext
	logger := r.logger.WithSpanContext(repoSpan.SpanContext())

	result, err := mapperID(
		ctx,
		r.tracer,
		r.logger,
		r.redis,
		deviceIDs,
		func(res string) string {
			return res
		},
		func(id int64) string {
			return r.redisPrefix + ":" + strconv.FormatInt(id, 10) + ":uid"
		},
		func(id int64) (*model.Device, error) {
			var device model.Device

			// avoid stale data in pgbouncer transaction mode
			tx := r.db.WithContext(ctx).Begin()
			defer tx.Rollback() // Rollback since it's a read-only transaction

			if err := tx.Select("id, uid").Where("id = ?", id).First(&device).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, domainErrors.ErrDeviceNotFound
				}
				return nil, err
			}
			return &device, nil
		},
		func(device *model.Device) string {
			return device.UID
		},
		r.redisMapDuration,
	)
	if err != nil {
		if errors.Is(err, domainErrors.ErrDeviceNotFound) {
			logger.Debug("Failed", map[string]interface{}{
				"error.message": err.Error(),
			})
		} else {
			logger.Error("error", map[string]interface{}{
				"error.message": err.Error(),
			})
		}
		repoSpan.AddEvent("Error", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
		return nil, err
	}

	repoSpan.AddEvent("success")

	return result, nil
}
