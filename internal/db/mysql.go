package db

import (
	"fmt"
	"pisa_server/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Manager struct {
	connections map[string]*gorm.DB
}

func NewManager(cfgs map[string]config.DBConfig) (*Manager, error) {
	m := &Manager{connections: make(map[string]*gorm.DB)}
	for name, cfg := range cfgs {
		db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("open db %s: %w", name, err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("get sql.DB %s: %w", name, err)
		}
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		m.connections[name] = db
	}
	return m, nil
}

func (m *Manager) Get(name string) *gorm.DB {
	return m.connections[name]
}

func (m *Manager) Close() {
	for _, db := range m.connections {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
