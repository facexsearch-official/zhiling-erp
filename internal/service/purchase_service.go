package service

import (
	"context"
	"fmt"
	"pisa_server/internal/model"
	customContext "pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/snowflake"
	"pisa_server/internal/repository"
	"time"

	"gorm.io/gorm"
)

type PurchaseService struct {
	db        *gorm.DB
	repo      *repository.PurchaseRepository
	itemRepo  *repository.PurchaseItemRepository
	goodsRepo *repository.GoodsRepository
}

func NewPurchaseService(
	dbConn *gorm.DB,
	repo *repository.PurchaseRepository,
	itemRepo *repository.PurchaseItemRepository,
	goodsRepo *repository.GoodsRepository,
) *PurchaseService {
	return &PurchaseService{
		db:        dbConn,
		repo:      repo,
		itemRepo:  itemRepo,
		goodsRepo: goodsRepo,
	}
}

func (s *PurchaseService) Create(ctx context.Context, purchase *model.Purchase, items []model.PurchaseItem) error {
	tenantID := customContext.GetTenantID(ctx)
	userID := customContext.GetUserID(ctx)

	purchase.ID = snowflake.GenID()
	purchase.OrderNo = s.generateOrderNo(ctx)
	purchase.TenantID = tenantID
	purchase.CreatedBy = userID
	purchase.Status = 1

	var total float64
	for i := range items {
		items[i].ID = snowflake.GenID()
		items[i].Amount = float64(items[i].Quantity) * items[i].UnitPrice
		items[i].TenantID = tenantID
		total += items[i].Amount
	}
	purchase.TotalAmount = total
	purchase.UnpaidAmount = total - purchase.PaidAmount

	if err := s.repo.Create(ctx, purchase); err != nil {
		return err
	}

	for i := range items {
		items[i].PurchaseID = purchase.ID
	}
	return s.itemRepo.BatchCreate(ctx, items)
}

func (s *PurchaseService) GetByID(ctx context.Context, id int64) (*model.Purchase, error) {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	items, err := s.itemRepo.ListByPurchaseID(ctx, id)
	if err != nil {
		return purchase, nil
	}
	purchase.Items = items
	return purchase, nil
}

func (s *PurchaseService) List(ctx context.Context, page, pageSize int, supplierID int64, status int8, keyword, dateFrom, dateTo string) ([]model.Purchase, int64) {
	return s.repo.List(ctx, page, pageSize, supplierID, status, keyword, dateFrom, dateTo)
}

func (s *PurchaseService) Audit(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	shopID := customContext.GetShopID(ctx)
	userID := customContext.GetUserID(ctx)

	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("进货单不存在")
	}
	if purchase.Status != 1 && purchase.Status != 2 {
		return fmt.Errorf("当前状态不允许审核")
	}

	items, err := s.itemRepo.ListByPurchaseID(ctx, purchase.ID)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("进货单无明细数据")
	}

	tx := s.db.Begin()

	for _, item := range items {
		var balance model.StockBalance
		result := tx.Where("tenant_id = ? AND shop_id = ? AND goods_id = ? AND warehouse_id = ?",
			tenantID, shopID, item.GoodsID, purchase.WarehouseID).First(&balance)

		beforeStock := 0
		if result.Error == gorm.ErrRecordNotFound {
			balance = model.StockBalance{
				TenantID:    tenantID,
				ShopID:      shopID,
				WarehouseID: purchase.WarehouseID,
				GoodsID:     item.GoodsID,
				Quantity:    0,
				CostPrice:   item.UnitPrice,
			}
		} else if result.Error == nil {
			beforeStock = balance.Quantity
		} else {
			tx.Rollback()
			return result.Error
		}

		balance.Quantity += item.Quantity
		balance.TotalCost = float64(balance.Quantity) * balance.CostPrice
		now := time.Now()
		balance.LastInTime = &now

		if err := tx.Save(&balance).Error; err != nil {
			tx.Rollback()
			return err
		}

		stockLog := map[string]interface{}{
			"id":           snowflake.GenID(),
			"tenant_id":    tenantID,
			"shop_id":      shopID,
			"warehouse_id": purchase.WarehouseID,
			"goods_id":     item.GoodsID,
			"type":         1,
			"quantity":     item.Quantity,
			"before_stock": beforeStock,
			"after_stock":  balance.Quantity,
			"related_type": "purchase",
			"related_id":   purchase.ID,
			"related_no":   purchase.OrderNo,
			"created_by":   userID,
			"created_at":   now,
		}
		if err := tx.Table("stock_logs").Create(stockLog).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	purchase.Status = 3
	if err := tx.Table("purchases").Save(purchase).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *PurchaseService) UnAudit(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	shopID := customContext.GetShopID(ctx)

	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("进货单不存在")
	}
	if purchase.Status != 3 {
		return fmt.Errorf("当前状态不允许反审核")
	}

	items, err := s.itemRepo.ListByPurchaseID(ctx, purchase.ID)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("进货单无明细数据")
	}

	tx := s.db.Begin()

	for _, item := range items {
		var balance model.StockBalance
		result := tx.Where("tenant_id = ? AND shop_id = ? AND goods_id = ? AND warehouse_id = ?",
			tenantID, shopID, item.GoodsID, purchase.WarehouseID).First(&balance)
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("库存记录不存在")
		}

		beforeStock := balance.Quantity
		balance.Quantity -= item.Quantity
		if balance.Quantity < 0 {
			tx.Rollback()
			return fmt.Errorf("库存不足，无法反审核")
		}
		balance.TotalCost = float64(balance.Quantity) * balance.CostPrice

		if err := tx.Save(&balance).Error; err != nil {
			tx.Rollback()
			return err
		}

		stockLog := map[string]interface{}{
			"id":           snowflake.GenID(),
			"tenant_id":    tenantID,
			"shop_id":      shopID,
			"warehouse_id": purchase.WarehouseID,
			"goods_id":     item.GoodsID,
			"type":         2,
			"quantity":     item.Quantity,
			"before_stock": beforeStock,
			"after_stock":  balance.Quantity,
			"related_type": "purchase_unaudit",
			"related_id":   purchase.ID,
			"related_no":   purchase.OrderNo,
			"created_at":   time.Now(),
		}
		if err := tx.Table("stock_logs").Create(stockLog).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	purchase.Status = 1
	if err := tx.Table("purchases").Save(purchase).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *PurchaseService) Delete(ctx context.Context, id int64) error {
	purchase, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("进货单不存在")
	}
	if purchase.Status == 3 {
		if err := s.UnAudit(ctx, id); err != nil {
			return err
		}
	}
	if err := s.itemRepo.DeleteByPurchaseID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *PurchaseService) generateOrderNo(ctx context.Context) string {
	today := time.Now().Format("20060102")
	var count int64
	s.repo.Scoped(ctx).Table("purchases").Where("order_no LIKE ?", "GH"+today+"%").Model(&model.Purchase{}).Count(&count)
	return fmt.Sprintf("GH%s%04d", today, count+1)
}
