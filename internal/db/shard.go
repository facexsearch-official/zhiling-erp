package db

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// ShardRouter 物理分表路由
type ShardRouter struct {
	shards   map[int]*gorm.DB // 0~N-1 → DB connection
	shardNum int
	tables   []string // 需要分片的表名列表
}

// ShardConfig 分片配置
type ShardConfig struct {
	ShardNum int                // 分片数量，如 16
	Tables   []string           // 需要分片的表名
	GetDB    func() *gorm.DB    // 获取数据库连接
}

// NewShardRouter 创建分片路由器
func NewShardRouter(cfg ShardConfig) *ShardRouter {
	r := &ShardRouter{
		shards:   make(map[int]*gorm.DB),
		shardNum: cfg.ShardNum,
		tables:   cfg.Tables,
	}
	db := cfg.GetDB()
	for i := 0; i < cfg.ShardNum; i++ {
		r.shards[i] = db // 所有分片共享同一个连接池，通过表名区分
	}
	return r
}

// GetShard 根据 tenant_id 计算分片号
func (r *ShardRouter) GetShard(tenantID int64) int {
	return int(tenantID % int64(r.shardNum))
}

// GetTable 获取分片后的表名：sales → sales_t11
func (r *ShardRouter) GetTable(tenantID int64, baseTable string) string {
	shard := r.GetShard(tenantID)
	return fmt.Sprintf("%s_t%02d", baseTable, shard)
}

// GetDB 获取分片对应的 DB（当前所有分片共享连接）
func (r *ShardRouter) GetDB(tenantID int64) *gorm.DB {
	shard := r.GetShard(tenantID)
	return r.shards[shard]
}

// ShardNum 获取分片数
func (r *ShardRouter) ShardNum() int {
	return r.shardNum
}

// Tables 获取需要分片的表列表
func (r *ShardRouter) Tables() []string {
	return r.tables
}

// AllTables 获取所有分片表名（用于跨分片 UNION ALL 查询）
func (r *ShardRouter) AllTables(baseTable string) []string {
	tables := make([]string, r.shardNum)
	for i := 0; i < r.shardNum; i++ {
		tables[i] = fmt.Sprintf("%s_t%02d", baseTable, i)
	}
	return tables
}

// ShardDB 分片数据库操作封装
type ShardDB struct {
	router *ShardRouter
	db     *gorm.DB
	mu     sync.RWMutex
}

// NewShardDB 创建分片数据库操作
func NewShardDB(router *ShardRouter, db *gorm.DB) *ShardDB {
	return &ShardDB{router: router, db: db}
}

// Table 获取分片表名
func (s *ShardDB) Table(tenantID int64, baseTable string) *gorm.DB {
	tableName := s.router.GetTable(tenantID, baseTable)
	return s.db.Table(tableName)
}

// Query 分片查询
func (s *ShardDB) Query(tenantID int64, baseTable string, dest interface{}, conditions ...interface{}) *gorm.DB {
	return s.Table(tenantID, baseTable).Where(conditions[0], conditions[1:]...).Find(dest)
}

// Exec 分片执行
func (s *ShardDB) Exec(tenantID int64, baseTable string, sql string, values ...interface{}) *gorm.DB {
	return s.Table(tenantID, baseTable).Exec(sql, values...)
}
