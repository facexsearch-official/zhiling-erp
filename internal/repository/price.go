package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"pisa_server/internal/model"

	customContext "pisa_server/internal/pkg/context"
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

// List 返回价格管理行（按货品分页，展开多单位/多规格）
func (r *PriceRepository) List(ctx context.Context, f PriceFilters, page, pageSize int) ([]PriceRow, int64, []string) {
	tenantID := customContext.GetTenantID(ctx)
	q := r.DB.Model(&model.Goods{}).Where("tenant_id = ?", tenantID)
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", kw, kw, kw)
	}
	if f.Spec != "" {
		q = q.Where("spec LIKE ?", "%"+f.Spec+"%")
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

	catMap := r.categoryMap(tenantID)
	unitMap := r.unitMap(tenantID)
	mainUnit := r.mainUnitMap(tenantID)

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
		rows = append(rows, expandGoods(g, catMap, unitMap, mainUnit)...)
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

func (r *PriceRepository) unitMap(tenantID int64) map[int64]string {
	var list []model.Unit
	r.DB.Where("tenant_id = ?", tenantID).Find(&list)
	m := map[int64]string{}
	for _, u := range list {
		m[u.ID] = u.Name
	}
	return m
}

func (r *PriceRepository) mainUnitMap(tenantID int64) map[int64]string {
	var list []model.GoodsUnit
	r.DB.Where("tenant_id = ?", tenantID).Order("is_main DESC, sort ASC, id ASC").Find(&list)
	m := map[int64]string{}
	for _, u := range list {
		if _, ok := m[u.GoodsID]; !ok && u.UnitName != "" {
			m[u.GoodsID] = u.UnitName
		}
	}
	return m
}

func expandGoods(g model.Goods, catMap, unitMap, mainUnit map[int64]string) []PriceRow {
	cat := ""
	if g.CategoryID != nil {
		cat = catMap[*g.CategoryID]
	}
	mainU := mainUnit[g.ID]
	if mainU == "" && g.UnitID != nil {
		mainU = unitMap[*g.UnitID]
	}
	specDisplay := g.Spec
	if specDisplay == "" {
		specDisplay = "--"
	}

	var outer map[string]map[string]map[string]interface{}
	if g.PriceRows != "" {
		_ = json.Unmarshal([]byte(g.PriceRows), &outer)
	}

	build := func(unitKey, specKey string, row map[string]interface{}) PriceRow {
		pr := PriceRow{
			GoodsID:  g.ID,
			UnitKey:  unitKey,
			SpecKey:  specKey,
			ImageURL: g.ImageURL,
			Name:     g.Name,
			Code:     g.Code,
			Barcode:  g.Barcode,
			Spec:     specDisplay,
			Category: cat,
			Unit:     mainU,
			Custom:   map[string]float64{},
			Status:   g.Status,
		}
		if row != nil {
			pr.Purchase = toFloat(row["purchase_price"])
			pr.Retail = toFloat(row["retail_price"])
			pr.Wholesale = toFloat(row["wholesale_price"])
			if v, ok := row["code"].(string); ok && v != "" {
				pr.Code = v
			}
			if v, ok := row["barcode"].(string); ok && v != "" {
				pr.Barcode = v
			}
			if cm, ok := row["custom"].(map[string]interface{}); ok {
				for k, v := range cm {
					pr.Custom[k] = toFloat(v)
				}
			}
			if d, ok := row["disabled"].(bool); ok {
				pr.Disabled = d
			}
		} else {
			pr.Purchase = g.PurchasePrice
			pr.Retail = g.RetailPrice
			pr.Wholesale = g.WholesalePrice
		}
		if unitKey != "" && unitKey != "__simple__" {
			pr.Unit = unitKey
		}
		if specKey != "" {
			pr.Spec = specKey
		}
		pr.Key = fmt.Sprintf("%d|%s|%s", g.ID, unitKey, specKey)
		return pr
	}

	if simple, ok := outer["__simple__"]; ok {
		return []PriceRow{build("__simple__", "", simple[""])}
	}
	if len(outer) > 0 {
		unitKeys := make([]string, 0, len(outer))
		for k := range outer {
			unitKeys = append(unitKeys, k)
		}
		sort.Strings(unitKeys)
		var rows []PriceRow
		for _, uk := range unitKeys {
			specs := outer[uk]
			specKeys := make([]string, 0, len(specs))
			for k := range specs {
				specKeys = append(specKeys, k)
			}
			sort.Strings(specKeys)
			for _, sk := range specKeys {
				rows = append(rows, build(uk, sk, specs[sk]))
			}
		}
		if len(rows) > 0 {
			return rows
		}
	}
	return []PriceRow{build("__simple__", "", nil)}
}

func adjustValue(v float64, ch PriceChange) float64 {
	if ch.Mode == "down" {
		return round2(v * (1 - ch.Percent/100))
	}
	return round2(v * (1 + ch.Percent/100))
}

func applyChange(row map[string]interface{}, ch PriceChange) {
	if strings.HasPrefix(ch.Field, "custom:") {
		col := strings.TrimPrefix(ch.Field, "custom:")
		cm, _ := row["custom"].(map[string]interface{})
		if cm == nil {
			cm = map[string]interface{}{}
		}
		cm[col] = adjustValue(toFloat(cm[col]), ch)
		row["custom"] = cm
		return
	}
	row[ch.Field] = adjustValue(toFloat(row[ch.Field]), ch)
}

// BatchAdjust 批量改价，返回修改的货品数
func (r *PriceRepository) BatchAdjust(ctx context.Context, scope string, keys []string,
	f PriceFilters, changes []PriceChange) (int, error) {
	tenantID := customContext.GetTenantID(ctx)

	var goodsList []model.Goods
	if scope == "selected" {
		ids := []int64{}
		seen := map[int64]bool{}
		for _, k := range keys {
			parts := strings.SplitN(k, "|", 2)
			id, _ := strconv.ParseInt(parts[0], 10, 64)
			if id > 0 && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return 0, nil
		}
		r.DB.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&goodsList)
	} else {
		q := r.DB.Model(&model.Goods{}).Where("tenant_id = ?", tenantID)
		if f.Keyword != "" {
			kw := "%" + f.Keyword + "%"
			q = q.Where("name LIKE ? OR code LIKE ? OR barcode LIKE ?", kw, kw, kw)
		}
		if f.Spec != "" {
			q = q.Where("spec LIKE ?", "%"+f.Spec+"%")
		}
		if f.Brand != "" && f.Brand != "全部" {
			q = q.Where("brand = ?", f.Brand)
		}
		if f.HideDisabled {
			q = q.Where("status = 1")
		}
		q.Find(&goodsList)
	}

	keySet := map[string]bool{}
	for _, k := range keys {
		keySet[k] = true
	}

	count := 0
	for _, g := range goodsList {
		var outer map[string]map[string]map[string]interface{}
		if g.PriceRows != "" {
			_ = json.Unmarshal([]byte(g.PriceRows), &outer)
		}
		if outer == nil {
			outer = map[string]map[string]map[string]interface{}{}
		}
		changed := false
		if simple, ok := outer["__simple__"]; ok {
			row := simple[""]
			if row == nil {
				row = map[string]interface{}{}
			}
			applyRowChanges(row, changes)
			simple[""] = row
			outer["__simple__"] = simple
			changed = true
			g.PurchasePrice = toFloat(row["purchase_price"])
			g.RetailPrice = toFloat(row["retail_price"])
			g.WholesalePrice = toFloat(row["wholesale_price"])
		} else {
			for uk, specs := range outer {
				for sk, row := range specs {
					key := fmt.Sprintf("%d|%s|%s", g.ID, uk, sk)
					if scope == "selected" && !keySet[key] {
						continue
					}
					if row == nil {
						row = map[string]interface{}{}
					}
					applyRowChanges(row, changes)
					specs[sk] = row
					changed = true
				}
			}
		}
		if !changed {
			continue
		}
		raw, _ := json.Marshal(outer)
		if err := r.DB.Model(&model.Goods{}).Where("id = ? AND tenant_id = ?", g.ID, tenantID).
			Updates(map[string]interface{}{
				"price_rows":      string(raw),
				"purchase_price":  g.PurchasePrice,
				"retail_price":    g.RetailPrice,
				"wholesale_price": g.WholesalePrice,
			}).Error; err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func applyRowChanges(row map[string]interface{}, changes []PriceChange) {
	for _, ch := range changes {
		if ch.Percent == 0 {
			continue
		}
		applyChange(row, ch)
	}
}

// UpdateRows 保存价格管理内联编辑的结果
func (r *PriceRepository) UpdateRows(ctx context.Context, updates []PriceRowUpdate) error {
	tenantID := customContext.GetTenantID(ctx)
	grouped := map[int64][]PriceRowUpdate{}
	for _, u := range updates {
		grouped[u.GoodsID] = append(grouped[u.GoodsID], u)
	}
	for goodsID, list := range grouped {
		var g model.Goods
		if err := r.DB.Where("id = ? AND tenant_id = ?", goodsID, tenantID).First(&g).Error; err != nil {
			continue
		}
		var outer map[string]map[string]map[string]interface{}
		if g.PriceRows != "" {
			_ = json.Unmarshal([]byte(g.PriceRows), &outer)
		}
		if outer == nil {
			outer = map[string]map[string]map[string]interface{}{}
		}
		for _, u := range list {
			uk := u.UnitKey
			if uk == "" {
				uk = "__simple__"
			}
			specs, ok := outer[uk]
			if !ok {
				specs = map[string]map[string]interface{}{}
			}
			row, ok := specs[u.SpecKey]
			if !ok || row == nil {
				row = map[string]interface{}{}
			}
			row["purchase_price"] = u.Purchase
			row["retail_price"] = u.Retail
			row["wholesale_price"] = u.Wholesale
			cm, _ := row["custom"].(map[string]interface{})
			if cm == nil {
				cm = map[string]interface{}{}
			}
			for k, v := range u.Custom {
				cm[k] = v
			}
			row["custom"] = cm
			specs[u.SpecKey] = row
			outer[uk] = specs
		}
		raw, _ := json.Marshal(outer)
		fields := map[string]interface{}{"price_rows": string(raw)}
		if simple, ok := outer["__simple__"]; ok {
			if row, ok := simple[""]; ok {
				fields["purchase_price"] = toFloat(row["purchase_price"])
				fields["retail_price"] = toFloat(row["retail_price"])
				fields["wholesale_price"] = toFloat(row["wholesale_price"])
			}
		}
		if err := r.DB.Model(&model.Goods{}).Where("id = ? AND tenant_id = ?", goodsID, tenantID).
			Updates(fields).Error; err != nil {
			return err
		}
	}
	return nil
}
