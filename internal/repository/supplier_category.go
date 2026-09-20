package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type SupplierCategoryRepository struct {
	BaseRepository
}

func NewSupplierCategoryRepository(base BaseRepository) *SupplierCategoryRepository {
	return &SupplierCategoryRepository{BaseRepository: base}
}

// ListAll 返回当前商户全部供应商分类
func (r *SupplierCategoryRepository) ListAll(ctx context.Context) ([]model.SupplierCategory, error) {
	var list []model.SupplierCategory
	err := r.Scoped(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// GetByID 查询单个分类
func (r *SupplierCategoryRepository) GetByID(ctx context.Context, id int64) (*model.SupplierCategory, error) {
	var cat model.SupplierCategory
	err := r.Scoped(ctx).Where("id = ?", id).First(&cat).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// Create 新建分类
func (r *SupplierCategoryRepository) Create(ctx context.Context, c *model.SupplierCategory) error {
	return r.Scoped(ctx).Create(c).Error
}

// Update 更新分类
func (r *SupplierCategoryRepository) Update(ctx context.Context, c *model.SupplierCategory) error {
	return r.Scoped(ctx).Save(c).Error
}

// Delete 删除分类
func (r *SupplierCategoryRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.SupplierCategory{}).Error
}

// HasSuppliers 判断分类下是否有供应商
func (r *SupplierCategoryRepository) HasSuppliers(ctx context.Context, categoryID int64) (bool, error) {
	var count int64
	err := r.Scoped(ctx).Model(&model.Supplier{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count > 0, err
}

// HasChildren 判断分类下是否有子分类
func (r *SupplierCategoryRepository) HasChildren(ctx context.Context, parentID int64) (bool, error) {
	var count int64
	err := r.Scoped(ctx).Model(&model.SupplierCategory{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

// GetDepth 计算分类层级深度（根节点=1）
func (r *SupplierCategoryRepository) GetDepth(ctx context.Context, categoryID int64) (int, error) {
	depth := 1
	currentID := categoryID
	for currentID != 0 {
		var cat model.SupplierCategory
		if err := r.Scoped(ctx).Where("id = ?", currentID).First(&cat).Error; err != nil {
			return depth, err
		}
		if cat.ParentID == 0 {
			break
		}
		depth++
		currentID = cat.ParentID
	}
	return depth, nil
}
