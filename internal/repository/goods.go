package repository

import (
	"context"
	"fmt"
	"time"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

type GoodsRepository struct {
	BaseRepository
}

func NewGoodsRepository(base BaseRepository) *GoodsRepository {
	return &GoodsRepository{BaseRepository: base}
}

func (r *GoodsRepository) List(ctx context.Context, page, pageSize int, keyword string, categoryID int64) ([]model.Goods, int64) {
	var total int64
	var list []model.Goods
	q := r.Scoped(ctx)
	if keyword != "" {
		q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	q.Model(&model.Goods{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	return list, total
}

func (r *GoodsRepository) ListAll(ctx context.Context) ([]model.Goods, error) {
	var list []model.Goods
	err := r.Scoped(ctx).Where("status = 1").Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *GoodsRepository) GetByID(ctx context.Context, id int64) (*model.Goods, error) {
	var goods model.Goods
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&goods).Error
	return &goods, err
}

// GetByIDWithChildren 返回货品详情并带出多单位 / 多规格
func (r *GoodsRepository) GetByIDWithChildren(ctx context.Context, id int64) (*model.Goods, error) {
	goods, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tenantID := customContext.GetTenantID(ctx)
	r.DB.Where("goods_id = ? AND tenant_id = ?", id, tenantID).Order("sort ASC, id ASC").Find(&goods.Units)
	r.DB.Where("goods_id = ? AND tenant_id = ?", id, tenantID).Order("sort ASC, id ASC").Find(&goods.Specs)
	return goods, nil
}

func (r *GoodsRepository) Create(ctx context.Context, goods *model.Goods) error {
	return r.Scoped(ctx).Create(goods).Error
}

// CreateWithChildren 事务创建货品 + 多单位 + 多规格
func (r *GoodsRepository) CreateWithChildren(ctx context.Context, goods *model.Goods,
	units []model.GoodsUnit, specs []model.GoodsSpec) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		goods.TenantID = tenantID
		if err := tx.Create(goods).Error; err != nil {
			return err
		}
		if err := insertChildren(tx, tenantID, goods.ID, units, specs); err != nil {
			return err
		}
		return nil
	})
}

func (r *GoodsRepository) Update(ctx context.Context, goods *model.Goods) error {
	return r.Scoped(ctx).Save(goods).Error
}

// UpdateWithChildren 事务更新货品并重建多单位 / 多规格
func (r *GoodsRepository) UpdateWithChildren(ctx context.Context, goods *model.Goods,
	units []model.GoodsUnit, specs []model.GoodsSpec) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 全字段更新主表（Select("*") 确保零值也能写入），排除子表 JSON
		if err := tx.Model(&model.Goods{}).Where("id = ? AND tenant_id = ?", goods.ID, tenantID).
			Select("*").Omit("id", "tenant_id", "created_at").Updates(goods).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ? AND tenant_id = ?", goods.ID, tenantID).
			Delete(&model.GoodsUnit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ? AND tenant_id = ?", goods.ID, tenantID).
			Delete(&model.GoodsSpec{}).Error; err != nil {
			return err
		}
		return insertChildren(tx, tenantID, goods.ID, units, specs)
	})
}

func insertChildren(tx *gorm.DB, tenantID, goodsID int64,
	units []model.GoodsUnit, specs []model.GoodsSpec) error {
	for i := range units {
		units[i].ID = 0
		units[i].TenantID = tenantID
		units[i].GoodsID = goodsID
		units[i].Sort = i
	}
	if len(units) > 0 {
		if err := tx.Create(&units).Error; err != nil {
			return err
		}
	}
	for i := range specs {
		specs[i].ID = 0
		specs[i].TenantID = tenantID
		specs[i].GoodsID = goodsID
		specs[i].Sort = i
	}
	if len(specs) > 0 {
		if err := tx.Create(&specs).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *GoodsRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Goods{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsUnit{}).Error; err != nil {
			return err
		}
		return tx.Where("goods_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsSpec{}).Error
	})
}

func (r *GoodsRepository) ListByIDs(ctx context.Context, ids []int64) ([]model.Goods, error) {
	var list []model.Goods
	tenantID := customContext.GetTenantID(ctx)
	err := r.DB.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&list).Error
	return list, err
}

// NextCode 生成下一个货品编号（形如 SP20260917001），带冲突兜底
func (r *GoodsRepository) NextCode(ctx context.Context) string {
	tenantID := customContext.GetTenantID(ctx)
	prefix := "SP" + time.Now().Format("20060102")
	var count int64
	r.DB.Model(&model.Goods{}).Where("tenant_id = ? AND code LIKE ?", tenantID, prefix+"%").
		Count(&count)
	for i := count + 1; i < count+1000; i++ {
		code := fmt.Sprintf("%s%04d", prefix, i)
		var n int64
		r.DB.Model(&model.Goods{}).Where("tenant_id = ? AND code = ?", tenantID, code).Count(&n)
		if n == 0 {
			return code
		}
	}
	return fmt.Sprintf("%s%04d", prefix, time.Now().Unix()%10000)
}
