package repository

import (
	"context"
	"encoding/json"
	"strings"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
)

// AttributeRepository 货品规格（辅助属性）字典
type AttributeRepository struct {
	BaseRepository
}

func NewAttributeRepository(base BaseRepository) *AttributeRepository {
	return &AttributeRepository{BaseRepository: base}
}

func (r *AttributeRepository) ListAll(ctx context.Context) ([]model.GoodsAttribute, error) {
	var list []model.GoodsAttribute
	err := r.Scoped(ctx).Where("status = 1").Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *AttributeRepository) List(ctx context.Context, name, content string) ([]model.GoodsAttribute, error) {
	var list []model.GoodsAttribute
	q := r.Scoped(ctx).Where("status = 1")
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if content != "" {
		q = q.Where("`values` LIKE ?", "%"+content+"%")
	}
	err := q.Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *AttributeRepository) GetByID(ctx context.Context, id int64) (*model.GoodsAttribute, error) {
	var a model.GoodsAttribute
	err := r.Scoped(ctx).Where("id = ?", id).First(&a).Error
	return &a, err
}

func (r *AttributeRepository) Create(ctx context.Context, a *model.GoodsAttribute) error {
	return r.Scoped(ctx).Create(a).Error
}

func (r *AttributeRepository) Save(ctx context.Context, a *model.GoodsAttribute) error {
	return r.Scoped(ctx).Save(a).Error
}

func (r *AttributeRepository) Delete(ctx context.Context, id int64) error {
	tenantID := customContext.GetTenantID(ctx)
	return r.DB.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.GoodsAttribute{}).Error
}

// HasGoods 判断规格是否被货品的 spec_groups 引用
func (r *AttributeRepository) HasGoods(ctx context.Context, name string) bool {
	if name == "" {
		return false
	}
	needle, _ := json.Marshal(map[string]string{"name": name})
	var n int64
	err := r.Scoped(ctx).Model(&model.Goods{}).
		Where("spec_groups IS NOT NULL AND spec_groups <> '' AND JSON_VALID(spec_groups) AND JSON_CONTAINS(spec_groups, ?)", string(needle)).
		Count(&n).Error
	if err != nil {
		// 退化方案：忽略 JSON 中的空白字符后做 LIKE
		like := "%\"name\":\"" + escapeLike(name) + "\"%"
		r.Scoped(ctx).Model(&model.Goods{}).
			Where("REPLACE(spec_groups, ' ', '') LIKE ?", like).
			Count(&n)
	}
	return n > 0
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
