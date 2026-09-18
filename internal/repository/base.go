package repository

import (
	"context"

	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

type BaseRepository struct {
	DB *gorm.DB
}

func NewBaseRepository(d *gorm.DB) BaseRepository {
	return BaseRepository{DB: d}
}

// Scoped 返回自动带 tenant_id 过滤的查询
func (r *BaseRepository) Scoped(ctx context.Context) *gorm.DB {
	return r.DB.Where("tenant_id = ?", customContext.GetTenantID(ctx))
}

// GetDB 返回原始 DB 连接
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.DB
}
