package repository

import (
	"context"
	"fmt"

	"pisa_server/internal/db"
	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

type BaseRepository struct {
	DB     *gorm.DB
	Router *db.ShardRouter
}

func NewBaseRepository(d *gorm.DB, router *db.ShardRouter) BaseRepository {
	return BaseRepository{DB: d, Router: router}
}

// Scoped 返回自动带 tenant_id 过滤的查询（不分片表）
func (r *BaseRepository) Scoped(ctx context.Context) *gorm.DB {
	return r.DB.Where("tenant_id = ?", customContext.GetTenantID(ctx))
}

// ScopedShard 返回分片表的查询（自动路由到正确分片）
func (r *BaseRepository) ScopedShard(ctx context.Context, baseTable string) *gorm.DB {
	tenantID := customContext.GetTenantID(ctx)
	tableName := r.Router.GetTable(tenantID, baseTable)
	return r.DB.Table(tableName).Where("tenant_id = ?", tenantID)
}

// GetShardTable 获取分片表名
func (r *BaseRepository) GetShardTable(ctx context.Context, baseTable string) string {
	return r.Router.GetTable(customContext.GetTenantID(ctx), baseTable)
}

// GetDB 返回原始 DB 连接
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.DB
}

// BuildUnion 构建跨分片 UNION ALL 查询（用于平台汇总）
func BuildUnion(db *gorm.DB, router *db.ShardRouter, baseTable, whereClause string, dest interface{}) error {
	var allSQL string
	for i := 0; i < router.ShardNum(); i++ {
		table := fmt.Sprintf("%s_t%02d", baseTable, i)
		if i > 0 {
			allSQL += " UNION ALL "
		}
		allSQL += fmt.Sprintf("SELECT * FROM %s WHERE %s", table, whereClause)
	}
	return db.Raw(allSQL).Scan(dest).Error
}
