package repository

import (
	"context"
	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

type CategoryRepository struct {
	BaseRepository
}

func NewCategoryRepository(base BaseRepository) *CategoryRepository {
	return &CategoryRepository{BaseRepository: base}
}

// ListAll 返回当前商户全部分类
func (r *CategoryRepository) ListAll(ctx context.Context) ([]model.GoodsCategory, error) {
	var list []model.GoodsCategory
	err := r.Scoped(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// GetByID 查询单个分类
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*model.GoodsCategory, error) {
	var cat model.GoodsCategory
	err := r.Scoped(ctx).Where("id = ?", id).First(&cat).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// Create 新建分类
func (r *CategoryRepository) Create(ctx context.Context, c *model.GoodsCategory) error {
	return r.Scoped(ctx).Create(c).Error
}

// Update 更新分类
func (r *CategoryRepository) Update(ctx context.Context, c *model.GoodsCategory) error {
	return r.Scoped(ctx).Save(c).Error
}

// Delete 删除分类
func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsCategory{}).Error
}

// HasGoods 判断分类下是否有商品
func (r *CategoryRepository) HasGoods(ctx context.Context, categoryID int64) (bool, error) {
	var count int64
	err := r.Scoped(ctx).Model(&model.Goods{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count > 0, err
}

// HasChildren 判断分类下是否有子分类
func (r *CategoryRepository) HasChildren(ctx context.Context, parentID int64) (bool, error) {
	var count int64
	err := r.Scoped(ctx).Model(&model.GoodsCategory{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count > 0, err
}

// GetDepth 计算分类层级深度（根节点=1）
func (r *CategoryRepository) GetDepth(ctx context.Context, categoryID int64) (int, error) {
	depth := 1
	currentID := categoryID
	for currentID != 0 {
		var cat model.GoodsCategory
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

// ExistsByName 同级下是否存在同名分类（排除自身 ID）
func (r *CategoryRepository) ExistsByName(ctx context.Context, name string, parentID, excludeID int64) (bool, error) {
	var count int64
	q := r.Scoped(ctx).Model(&model.GoodsCategory{}).Where("name = ? AND parent_id = ?", name, parentID)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}
