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

func (r *GoodsRepository) List(ctx context.Context, page, pageSize int, keyword string, categoryID int64, hideDisabled, hideZero bool) ([]model.Goods, int64) {
	var total int64
	var list []model.Goods
	q := r.Scoped(ctx)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ? OR EXISTS (SELECT 1 FROM goods_prices gp WHERE gp.goods_id = goods.id AND gp.barcode LIKE ?)",
			kw, kw, kw, kw)
	}
	if hideDisabled {
		q = q.Where("status = 1")
	}
	if hideZero {
		q = q.Where("current_stock <> 0")
	}
	if categoryID == -1 {
		// 未分类：无分类或分类为 0
		q = q.Where("category_id IS NULL OR category_id = 0")
	} else if categoryID > 0 {
		sub := r.DB.Model(&model.GoodsCategory{}).Select("id").Where("parent_id = ?", categoryID)
		if tid := customContext.GetTenantID(ctx); tid > 0 {
			sub = sub.Where("tenant_id = ?", tid)
		}
		q = q.Where("category_id = ? OR category_id IN (?)", categoryID, sub)
	}
	q.Model(&model.Goods{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list)
	r.attachSummaries(ctx, list)
	return list, total
}

// attachSummaries 批量加载单位与价格聚合，填充列表摘要
func (r *GoodsRepository) attachSummaries(ctx context.Context, list []model.Goods) {
	if len(list) == 0 {
		return
	}
	tenantID := customContext.GetTenantID(ctx)
	ids := make([]int64, 0, len(list))
	for _, g := range list {
		ids = append(ids, g.ID)
	}

	var units []model.GoodsUnit
	r.DB.Where("tenant_id = ? AND goods_id IN ?", tenantID, ids).
		Order("is_main DESC, sort ASC, id ASC").Find(&units)
	unitMap := map[int64][]model.GoodsUnit{}
	for _, u := range units {
		unitMap[u.GoodsID] = append(unitMap[u.GoodsID], u)
	}

	type aggRow struct {
		GoodsID      int64   `gorm:"column:goods_id"`
		UnitKey      string  `gorm:"column:unit_key"`
		RetailMin    float64 `gorm:"column:retail_min"`
		RetailMax    float64 `gorm:"column:retail_max"`
		WholesaleMin float64 `gorm:"column:wholesale_min"`
		WholesaleMax float64 `gorm:"column:wholesale_max"`
		Codes        string  `gorm:"column:codes"`
	}
	var aggs []aggRow
	r.DB.Model(&model.GoodsPrice{}).
		Select("goods_id, unit_key, MIN(retail_price) AS retail_min, MAX(retail_price) AS retail_max, "+
			"MIN(wholesale_price) AS wholesale_min, MAX(wholesale_price) AS wholesale_max, "+
			"GROUP_CONCAT(DISTINCT code SEPARATOR '/') AS codes").
		Where("tenant_id = ? AND goods_id IN ?", tenantID, ids).
		Group("goods_id, unit_key").Scan(&aggs)
	aggMap := map[string]aggRow{}
	for _, a := range aggs {
		aggMap[fmt.Sprintf("%d|%s", a.GoodsID, a.UnitKey)] = a
	}

	for i := range list {
		g := &list[i]
		us := unitMap[g.ID]
		if len(us) == 0 {
			// 无多单位：用主单位名兜底
			us = []model.GoodsUnit{{UnitName: g.MainUnit, Factor: 1, IsMain: 1}}
		}
		summaries := make([]model.GoodsUnitSummary, 0, len(us))
		for _, u := range us {
			a := aggMap[fmt.Sprintf("%d|%s", g.ID, u.UnitName)]
			summaries = append(summaries, model.GoodsUnitSummary{
				UnitName:     u.UnitName,
				Factor:       u.Factor,
				IsMain:       u.IsMain,
				RetailMin:    a.RetailMin,
				RetailMax:    a.RetailMax,
				WholesaleMin: a.WholesaleMin,
				WholesaleMax: a.WholesaleMax,
				Codes:        a.Codes,
			})
		}
		g.UnitSummaries = summaries
	}
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

// GetByIDWithChildren 返回货品详情并带出多单位 / 价格 / 库存
func (r *GoodsRepository) GetByIDWithChildren(ctx context.Context, id int64) (*model.Goods, error) {
	goods, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tenantID := customContext.GetTenantID(ctx)
	r.DB.Where("goods_id = ? AND tenant_id = ?", id, tenantID).
		Order("is_main DESC, sort ASC, id ASC").Find(&goods.Units)
	r.DB.Where("goods_id = ? AND tenant_id = ?", id, tenantID).
		Order("sort ASC, id ASC").Find(&goods.Prices)
	r.DB.Where("goods_id = ? AND tenant_id = ?", id, tenantID).
		Order("id ASC").Find(&goods.Stocks)
	return goods, nil
}

func (r *GoodsRepository) Create(ctx context.Context, goods *model.Goods) error {
	return r.Scoped(ctx).Create(goods).Error
}

// CreateWithChildren 事务创建货品 + 多单位 + 价格明细 + 库存
func (r *GoodsRepository) CreateWithChildren(ctx context.Context, goods *model.Goods,
	units []model.GoodsUnit, prices []model.GoodsPrice, stocks []model.GoodsStock) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		goods.TenantID = tenantID
		if err := tx.Create(goods).Error; err != nil {
			return err
		}
		if err := insertChildren(tx, tenantID, goods.ID, units, prices, stocks); err != nil {
			return err
		}
		return recomputeAggregate(tx, tenantID, goods.ID)
	})
}

func (r *GoodsRepository) Update(ctx context.Context, goods *model.Goods) error {
	return r.Scoped(ctx).Save(goods).Error
}

// UpdateWithChildren 事务更新货品并重建多单位 / 价格 / 库存
func (r *GoodsRepository) UpdateWithChildren(ctx context.Context, goods *model.Goods,
	units []model.GoodsUnit, prices []model.GoodsPrice, stocks []model.GoodsStock) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 全字段更新主表（Select("*") 确保零值也能写入）
		if err := tx.Model(&model.Goods{}).Where("id = ? AND tenant_id = ?", goods.ID, tenantID).
			Select("*").Omit("id", "tenant_id", "created_at").Updates(goods).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ? AND tenant_id = ?", goods.ID, tenantID).
			Delete(&model.GoodsUnit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ? AND tenant_id = ?", goods.ID, tenantID).
			Delete(&model.GoodsPrice{}).Error; err != nil {
			return err
		}
		if err := tx.Where("goods_id = ? AND tenant_id = ?", goods.ID, tenantID).
			Delete(&model.GoodsStock{}).Error; err != nil {
			return err
		}
		if err := insertChildren(tx, tenantID, goods.ID, units, prices, stocks); err != nil {
			return err
		}
		return recomputeAggregate(tx, tenantID, goods.ID)
	})
}

func insertChildren(tx *gorm.DB, tenantID, goodsID int64,
	units []model.GoodsUnit, prices []model.GoodsPrice, stocks []model.GoodsStock) error {
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
	for i := range prices {
		prices[i].ID = 0
		prices[i].TenantID = tenantID
		prices[i].GoodsID = goodsID
		prices[i].Sort = i
	}
	if len(prices) > 0 {
		if err := tx.Create(&prices).Error; err != nil {
			return err
		}
	}
	for i := range stocks {
		stocks[i].ID = 0
		stocks[i].TenantID = tenantID
		stocks[i].GoodsID = goodsID
	}
	if len(stocks) > 0 {
		if err := tx.Create(&stocks).Error; err != nil {
			return err
		}
	}
	return nil
}

// recomputeAggregate 重算货品冗余聚合列（主单位 / 零售价范围 / 总库存）
func recomputeAggregate(tx *gorm.DB, tenantID, goodsID int64) error {
	var mainUnit string
	tx.Model(&model.GoodsUnit{}).
		Where("goods_id = ? AND tenant_id = ?", goodsID, tenantID).
		Order("is_main DESC, sort ASC, id ASC").Limit(1).Pluck("unit_name", &mainUnit)

	var pr struct {
		RMin float64 `gorm:"column:rmin"`
		RMax float64 `gorm:"column:rmax"`
	}
	tx.Model(&model.GoodsPrice{}).
		Where("goods_id = ? AND tenant_id = ? AND unit_key = ? AND disabled = 0", goodsID, tenantID, mainUnit).
		Select("COALESCE(MIN(retail_price),0) AS rmin, COALESCE(MAX(retail_price),0) AS rmax").
		Scan(&pr)

	var total int
	tx.Model(&model.GoodsStock{}).
		Where("goods_id = ? AND tenant_id = ?", goodsID, tenantID).
		Select("COALESCE(SUM(stock),0)").Scan(&total)

	// 主单位首行价格同步到主表（兼容开单等模块读取）
	var first model.GoodsPrice
	tx.Where("goods_id = ? AND tenant_id = ? AND unit_key = ?", goodsID, tenantID, mainUnit).
		Order("sort ASC, id ASC").First(&first)

	return tx.Model(&model.Goods{}).Where("id = ? AND tenant_id = ?", goodsID, tenantID).
		Updates(map[string]interface{}{
			"main_unit":       mainUnit,
			"retail_min":      pr.RMin,
			"retail_max":      pr.RMax,
			"total_stock":     total,
			"purchase_price":  first.PurchasePrice,
			"retail_price":    first.RetailPrice,
			"wholesale_price": first.WholesalePrice,
		}).Error
}

// RecomputeGoodsAggregate 供价格模块在改价后同步聚合列
func RecomputeGoodsAggregate(tx *gorm.DB, tenantID, goodsID int64) error {
	return recomputeAggregate(tx, tenantID, goodsID)
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
		if err := tx.Where("goods_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsPrice{}).Error; err != nil {
			return err
		}
		return tx.Where("goods_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsStock{}).Error
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
