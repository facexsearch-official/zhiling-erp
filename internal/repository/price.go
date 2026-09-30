package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"

	"gorm.io/gorm"
)

// PriceRepository 价格管理
type PriceRepository struct {
	BaseRepository
}

func NewPriceRepository(base BaseRepository) *PriceRepository {
	return &PriceRepository{BaseRepository: base}
}

// PriceRow 价格管理列表行
type PriceRow struct {
	GoodsID   int64              `json:"goods_id"`
	Key       string             `json:"key"`
	UnitKey   string             `json:"unit_key"`
	SpecKey   string             `json:"spec_key"`
	ImageURL  string             `json:"image_url"`
	Name      string             `json:"name"`
	Code      string             `json:"code"`
	Barcode   string             `json:"barcode"`
	Spec      string             `json:"spec"`
	Category  string             `json:"category"`
	Unit      string             `json:"unit"`
	Purchase  float64            `json:"purchase_price"`
	Retail    float64            `json:"retail_price"`
	Wholesale float64            `json:"wholesale_price"`
	Custom    map[string]float64 `json:"custom"`
	Disabled  bool               `json:"disabled"`
	Status    int8               `json:"status"`
}

// PriceFilters 价格列表过滤条件
type PriceFilters struct {
	Keyword      string
	Spec         string
	Brand        string
	HideDisabled bool
}

// PriceChange 批量改价项
type PriceChange struct {
	Field   string  `json:"field"`
	Mode    string  `json:"mode"`
	Percent float64 `json:"percent"`
}

// PriceRowUpdate 单价行显式更新
type PriceRowUpdate struct {
	GoodsID   int64              `json:"goods_id"`
	UnitKey   string             `json:"unit_key"`
	SpecKey   string             `json:"spec_key"`
	Purchase  float64            `json:"purchase_price"`
	Retail    float64            `json:"retail_price"`
	Wholesale float64            `json:"wholesale_price"`
	Custom    map[string]float64 `json:"custom"`
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	}
	return 0
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5*sign(v))) / 100
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

func parseStringArray(raw string) []string {
	if raw == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return nil
	}
	return arr
}

func parseCustom(raw string) map[string]float64 {
	m := map[string]float64{}
	if raw == "" {
		return m
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return m
	}
	for k, v := range obj {
		m[k] = toFloat(v)
	}
	return m
}

// List 返回价格管理行（按货品分页，展开多单位/多规格）
func (r *PriceRepository) List(ctx context.Context, f PriceFilters, page, pageSize int) ([]PriceRow, int64, []string) {
	tenantID := customContext.GetTenantID(ctx)
	q := r.DB.Model(&model.Goods{}).Where("tenant_id = ?", tenantID)
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("(name LIKE ? OR code LIKE ? OR EXISTS (SELECT 1 FROM goods_prices gp WHERE gp.goods_id = goods.id AND (gp.code LIKE ? OR gp.barcode LIKE ?)))",
			kw, kw, kw, kw)
	}
	if f.Spec != "" {
		q = q.Where("EXISTS (SELECT 1 FROM goods_prices gp WHERE gp.goods_id = goods.id AND gp.spec_key LIKE ?)", "%"+f.Spec+"%")
	}
	if f.Brand != "" && f.Brand != "全部" {
		q = q.Where("brand = ?", f.Brand)
	}
	if f.HideDisabled {
		q = q.Where("status = 1")
	}
	var total int64
	q.Count(&total)
	var goodsList []model.Goods
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&goodsList)

	ids := make([]int64, 0, len(goodsList))
	for _, g := range goodsList {
		ids = append(ids, g.ID)
	}
	priceMap := map[int64][]model.GoodsPrice{}
	if len(ids) > 0 {
		var prices []model.GoodsPrice
		r.DB.Where("tenant_id = ? AND goods_id IN ?", tenantID, ids).
			Order("goods_id ASC, sort ASC, id ASC").Find(&prices)
		for _, p := range prices {
			priceMap[p.GoodsID] = append(priceMap[p.GoodsID], p)
		}
	}

	catMap := r.categoryMap(tenantID)

	rows := []PriceRow{}
	colSet := map[string]bool{}
	cols := []string{}
	for _, g := range goodsList {
		for _, c := range parseStringArray(g.PriceColumns) {
			if !colSet[c] {
				colSet[c] = true
				cols = append(cols, c)
			}
		}
		rows = append(rows, expandGoodsRows(g, priceMap[g.ID], catMap)...)
	}
	return rows, total, cols
}

func (r *PriceRepository) categoryMap(tenantID int64) map[int64]string {
	var list []model.GoodsCategory
	r.DB.Where("tenant_id = ?", tenantID).Find(&list)
	m := map[int64]string{}
	for _, c := range list {
		m[c.ID] = c.Name
	}
	return m
}

func expandGoodsRows(g model.Goods, prices []model.GoodsPrice, catMap map[int64]string) []PriceRow {
	cat := ""
	if g.CategoryID != nil {
		cat = catMap[int64(*g.CategoryID)]
	}
	if len(prices) == 0 {
		return []PriceRow{buildPriceRow(g, model.GoodsPrice{}, cat)}
	}
	rows := make([]PriceRow, 0, len(prices))
	for _, p := range prices {
		rows = append(rows, buildPriceRow(g, p, cat))
	}
	return rows
}

func buildPriceRow(g model.Goods, p model.GoodsPrice, cat string) PriceRow {
	unit := p.UnitKey
	if unit == "" || unit == "__simple__" {
		unit = g.MainUnit
	}
	spec := p.SpecKey
	if spec == "" {
		spec = g.Spec
		if spec == "" {
			spec = "--"
		}
	}
	code := p.Code
	if code == "" {
		code = g.Code
	}
	barcode := p.Barcode
	if barcode == "" {
		barcode = g.Barcode
	}
	return PriceRow{
		GoodsID:   g.ID,
		Key:       fmt.Sprintf("%d|%s|%s", g.ID, p.UnitKey, p.SpecKey),
		UnitKey:   p.UnitKey,
		SpecKey:   p.SpecKey,
		ImageURL:  g.ImageURL,
		Name:      g.Name,
		Code:      code,
		Barcode:   barcode,
		Spec:      spec,
		Category:  cat,
		Unit:      unit,
		Purchase:  p.PurchasePrice,
		Retail:    p.RetailPrice,
		Wholesale: p.WholesalePrice,
		Custom:    parseCustom(p.Custom),
		Disabled:  p.Disabled != 0,
		Status:    g.Status,
	}
}

func adjustValue(v float64, ch PriceChange) float64 {
	if ch.Mode == "down" {
		return round2(v * (1 - ch.Percent/100))
	}
	return round2(v * (1 + ch.Percent/100))
}

func applyPriceChange(p *model.GoodsPrice, ch PriceChange) {
	if strings.HasPrefix(ch.Field, "custom:") {
		col := strings.TrimPrefix(ch.Field, "custom:")
		cm := map[string]interface{}{}
		if p.Custom != "" {
			_ = json.Unmarshal([]byte(p.Custom), &cm)
		}
		cm[col] = adjustValue(toFloat(cm[col]), ch)
		raw, _ := json.Marshal(cm)
		p.Custom = string(raw)
		return
	}
	switch ch.Field {
	case "purchase_price":
		p.PurchasePrice = adjustValue(p.PurchasePrice, ch)
	case "retail_price":
		p.RetailPrice = adjustValue(p.RetailPrice, ch)
	case "wholesale_price":
		p.WholesalePrice = adjustValue(p.WholesalePrice, ch)
	}
}

// BatchAdjust 批量改价，返回修改的货品数
func (r *PriceRepository) BatchAdjust(ctx context.Context, scope string, keys []string,
	f PriceFilters, changes []PriceChange) (int, error) {
	tenantID := customContext.GetTenantID(ctx)

	var goodsIDs []int64
	if scope == "selected" {
		seen := map[int64]bool{}
		for _, k := range keys {
			parts := strings.SplitN(k, "|", 2)
			id, _ := strconv.ParseInt(parts[0], 10, 64)
			if id > 0 && !seen[id] {
				seen[id] = true
				goodsIDs = append(goodsIDs, id)
			}
		}
		if len(goodsIDs) == 0 {
			return 0, nil
		}
	} else {
		q := r.DB.Model(&model.Goods{}).Where("tenant_id = ?", tenantID)
		if f.Keyword != "" {
			kw := "%" + f.Keyword + "%"
			q = q.Where("(name LIKE ? OR code LIKE ? OR EXISTS (SELECT 1 FROM goods_prices gp WHERE gp.goods_id = goods.id AND (gp.code LIKE ? OR gp.barcode LIKE ?)))",
				kw, kw, kw, kw)
		}
		if f.Spec != "" {
			q = q.Where("EXISTS (SELECT 1 FROM goods_prices gp WHERE gp.goods_id = goods.id AND gp.spec_key LIKE ?)", "%"+f.Spec+"%")
		}
		if f.Brand != "" && f.Brand != "全部" {
			q = q.Where("brand = ?", f.Brand)
		}
		if f.HideDisabled {
			q = q.Where("status = 1")
		}
		var list []model.Goods
		q.Find(&list)
		for _, g := range list {
			goodsIDs = append(goodsIDs, g.ID)
		}
	}

	keySet := map[string]bool{}
	for _, k := range keys {
		keySet[k] = true
	}

	count := 0
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		for _, gid := range goodsIDs {
			var prices []model.GoodsPrice
			if err := tx.Where("goods_id = ? AND tenant_id = ?", gid, tenantID).Find(&prices).Error; err != nil {
				return err
			}
			changed := false
			for i := range prices {
				p := &prices[i]
				key := fmt.Sprintf("%d|%s|%s", gid, p.UnitKey, p.SpecKey)
				if scope == "selected" && !keySet[key] {
					continue
				}
				applyRowChanges(p, changes)
				if err := tx.Model(&model.GoodsPrice{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
					"purchase_price":  p.PurchasePrice,
					"retail_price":    p.RetailPrice,
					"wholesale_price": p.WholesalePrice,
					"custom":          p.Custom,
				}).Error; err != nil {
					return err
				}
				changed = true
			}
			if changed {
				if err := RecomputeGoodsAggregate(tx, tenantID, gid); err != nil {
					return err
				}
				count++
			}
		}
		return nil
	})
	return count, err
}

func applyRowChanges(p *model.GoodsPrice, changes []PriceChange) {
	for _, ch := range changes {
		if ch.Percent == 0 {
			continue
		}
		applyPriceChange(p, ch)
	}
}

// UpdateRows 保存价格管理内联编辑的结果
func (r *PriceRepository) UpdateRows(ctx context.Context, updates []PriceRowUpdate) error {
	tenantID := customContext.GetTenantID(ctx)
	grouped := map[int64][]PriceRowUpdate{}
	for _, u := range updates {
		grouped[u.GoodsID] = append(grouped[u.GoodsID], u)
	}
	return r.DB.Transaction(func(tx *gorm.DB) error {
		for goodsID, list := range grouped {
			for _, u := range list {
				uk := u.UnitKey
				if uk == "__simple__" {
					uk = ""
				}
				var p model.GoodsPrice
				err := tx.Where("goods_id = ? AND tenant_id = ? AND unit_key = ? AND spec_key = ?",
					goodsID, tenantID, uk, u.SpecKey).First(&p).Error
				if err != nil {
					p = model.GoodsPrice{TenantID: tenantID, GoodsID: goodsID, UnitKey: uk, SpecKey: u.SpecKey}
				}
				p.PurchasePrice = u.Purchase
				p.RetailPrice = u.Retail
				p.WholesalePrice = u.Wholesale
				cm := map[string]interface{}{}
				if p.Custom != "" {
					_ = json.Unmarshal([]byte(p.Custom), &cm)
				}
				for k, v := range u.Custom {
					cm[k] = v
				}
				raw, _ := json.Marshal(cm)
				p.Custom = string(raw)
				if p.ID == 0 {
					if err := tx.Create(&p).Error; err != nil {
						return err
					}
				} else {
					if err := tx.Model(&model.GoodsPrice{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
						"purchase_price":  p.PurchasePrice,
						"retail_price":    p.RetailPrice,
						"wholesale_price": p.WholesalePrice,
						"custom":          p.Custom,
					}).Error; err != nil {
						return err
					}
				}
			}
			if err := RecomputeGoodsAggregate(tx, tenantID, goodsID); err != nil {
				return err
			}
		}
		return nil
	})
}
