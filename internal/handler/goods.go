package handler

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/repository"

	"github.com/gin-gonic/gin"
)

type GoodsHandler struct {
	repo *repository.GoodsRepository
}

func NewGoodsHandler(repo *repository.GoodsRepository) *GoodsHandler {
	return &GoodsHandler{repo: repo}
}

// goodsPayload 接收前端 payload，额外承载 JSON 字符串形式的 price_rows / stock_rows
type goodsPayload struct {
	model.Goods
	PriceRowsRaw string `json:"price_rows"`
	StockRowsRaw string `json:"stock_rows"`
}

func (h *GoodsHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	categoryID, _ := strconv.ParseInt(c.Query("category_id"), 10, 64)
	hideDisabled := c.Query("hide_disabled") == "1"
	hideZero := c.Query("hide_zero") == "1"
	list, total := h.repo.List(ctx, page, pageSize, keyword, categoryID, hideDisabled, hideZero)
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
	}
	response.OKPage(c, list, total, page, pageSize)
}

type goodsBatchFilters struct {
	Keyword      string `json:"keyword"`
	CategoryID   int64  `json:"category_id"`
	HideDisabled bool   `json:"hide_disabled"`
	HideZero     bool   `json:"hide_zero"`
}

type goodsBatchReq struct {
	Action  string            `json:"action"` // update | delete
	Scope   string            `json:"scope"`  // selected | query
	IDs     []string          `json:"ids"`
	Filters goodsBatchFilters `json:"filters"`
	Patch   struct {
		Status           *int8    `json:"status"`
		CategoryID       *string  `json:"category_id"`
		RetailPrice      *float64 `json:"retail_price"`
		WholesalePrice   *float64 `json:"wholesale_price"`
		PurchasePrice    *float64 `json:"purchase_price"`
		Brand            *string  `json:"brand"`
		Origin           *string  `json:"origin"`
		Remark           *string  `json:"remark"`
		Images           *string  `json:"images"`
		EnableStockAlert *int8    `json:"enable_stock_alert"`
		MinStock         *int     `json:"min_stock"`
		SafetyStock      *int     `json:"safety_stock"`
		MaxStock         *int     `json:"max_stock"`
		HasBatch         *int8    `json:"has_batch"`
		HasShelfLife     *int8    `json:"has_shelf_life"`
		ShelfLifeDays    *int     `json:"shelf_life_days"`
		ExpiryAlert      *int8    `json:"expiry_alert"`
		ExpiryWarnDays   *int     `json:"expiry_warn_days"`
		HasSerial        *int8    `json:"has_serial"`
	} `json:"patch"`
}

// BatchUpdate 批量修改/删除商品（范围：选中 或 当前查询条件）
func (h *GoodsHandler) BatchUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req goodsBatchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	q := h.repo.DB.Table("goods").Where("tenant_id = ?", tenantID)
	if req.Scope == "query" {
		f := req.Filters
		if f.Keyword != "" {
			kw := "%" + f.Keyword + "%"
			q = q.Where("(name LIKE ? OR code LIKE ? OR barcode LIKE ?)", kw, kw, kw)
		}
		if f.CategoryID == -1 {
			q = q.Where("category_id IS NULL OR category_id = 0")
		} else if f.CategoryID > 0 {
			q = q.Where("category_id = ?", f.CategoryID)
		}
		if f.HideDisabled {
			q = q.Where("status = 1")
		}
		if f.HideZero {
			q = q.Where("current_stock <> 0")
		}
	} else {
		ids := make([]int64, 0, len(req.IDs))
		for _, s := range req.IDs {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
				ids = append(ids, n)
			}
		}
		if len(ids) == 0 {
			response.BadRequest(c, "请选择商品")
			return
		}
		q = q.Where("id IN ?", ids)
	}

	if req.Action == "delete" {
		res := q.Delete(&model.Goods{})
		if res.Error != nil {
			response.ServerError(c, "批量删除失败")
			return
		}
		response.OK(c, gin.H{"count": res.RowsAffected})
		return
	}

	upd := map[string]interface{}{}
	p := req.Patch
	if p.Status != nil {
		upd["status"] = *p.Status
	}
	if p.CategoryID != nil {
		if *p.CategoryID == "" || *p.CategoryID == "0" {
			upd["category_id"] = nil
		} else if n, err := strconv.ParseInt(*p.CategoryID, 10, 64); err == nil {
			upd["category_id"] = n
		}
	}
	if p.RetailPrice != nil {
		upd["retail_price"] = *p.RetailPrice
	}
	if p.WholesalePrice != nil {
		upd["wholesale_price"] = *p.WholesalePrice
	}
	if p.PurchasePrice != nil {
		upd["purchase_price"] = *p.PurchasePrice
	}
	if p.Brand != nil {
		upd["brand"] = *p.Brand
	}
	if p.Origin != nil {
		upd["origin"] = *p.Origin
	}
	if p.Remark != nil {
		upd["remark"] = *p.Remark
	}
	if p.Images != nil {
		upd["images"] = *p.Images
	}
	if p.EnableStockAlert != nil {
		upd["enable_stock_alert"] = *p.EnableStockAlert
	}
	if p.MinStock != nil {
		upd["min_stock"] = *p.MinStock
	}
	if p.SafetyStock != nil {
		upd["safety_stock"] = *p.SafetyStock
	}
	if p.MaxStock != nil {
		upd["max_stock"] = *p.MaxStock
	}
	if p.HasBatch != nil {
		upd["has_batch"] = *p.HasBatch
	}
	if p.HasShelfLife != nil {
		upd["has_shelf_life"] = *p.HasShelfLife
	}
	if p.ShelfLifeDays != nil {
		upd["shelf_life_days"] = *p.ShelfLifeDays
	}
	if p.ExpiryAlert != nil {
		upd["expiry_alert"] = *p.ExpiryAlert
	}
	if p.ExpiryWarnDays != nil {
		upd["expiry_warn_days"] = *p.ExpiryWarnDays
	}
	if p.HasSerial != nil {
		upd["has_serial"] = *p.HasSerial
	}
	if len(upd) == 0 {
		response.BadRequest(c, "没有需要修改的内容")
		return
	}
	res := q.Updates(upd)
	if res.Error != nil {
		response.ServerError(c, "批量修改失败")
		return
	}
	response.OK(c, gin.H{"count": res.RowsAffected})
}

func (h *GoodsHandler) ListAll(c *gin.Context) {
	ctx := c.Request.Context()
	list, err := h.repo.ListAll(ctx)
	if err != nil {
		response.ServerError(c, "查询货品失败")
		return
	}
	response.OK(c, list)
}

// GetByID 返回货品详情（含多单位 / 价格 / 库存）
func (h *GoodsHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	goods, err := h.repo.GetByIDWithChildren(ctx, id)
	if err != nil {
		response.NotFound(c, "货品不存在")
		return
	}
	response.OK(c, goods)
}

// NextCode 生成下一个货品编号
func (h *GoodsHandler) NextCode(c *gin.Context) {
	response.OK(c, gin.H{"code": h.repo.NextCode(c.Request.Context())})
}

// Brands 返回当前商户所有不重复的品牌
func (h *GoodsHandler) Brands(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var list []string
	h.repo.DB.Model(&model.Goods{}).
		Where("tenant_id = ? AND brand != '' AND brand IS NOT NULL", tenantID).
		Distinct("brand").Pluck("brand", &list)
	response.OK(c, list)
}

// Origins 返回当前商户所有不重复的产地
func (h *GoodsHandler) Origins(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var list []string
	h.repo.DB.Model(&model.Goods{}).
		Where("tenant_id = ? AND origin != '' AND origin IS NOT NULL", tenantID).
		Distinct("origin").Pluck("origin", &list)
	response.OK(c, list)
}

func (h *GoodsHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var payload goodsPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if strings.TrimSpace(payload.Name) == "" {
		response.BadRequest(c, "货品名称不能为空")
		return
	}
	if utf8.RuneCountInString(payload.Brand) > 100 {
		response.BadRequest(c, "品牌名称不能超过100个字符")
		return
	}
	goods := payload.Goods
	goods.ID = 0
	goods.Status = 1
	goods.TenantID = context.GetTenantID(ctx)
	units, prices, stocks := h.buildChildren(&goods, payload.PriceRowsRaw, payload.StockRowsRaw)
	if err := h.repo.CreateWithChildren(ctx, &goods, units, prices, stocks); err != nil {
		response.ServerError(c, "创建货品失败")
		return
	}
	h.respondWithDetail(c, goods.ID, goods)
}

func (h *GoodsHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		response.NotFound(c, "货品不存在")
		return
	}
	payload := goodsPayload{Goods: *existing}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if strings.TrimSpace(payload.Name) == "" {
		response.BadRequest(c, "货品名称不能为空")
		return
	}
	if utf8.RuneCountInString(payload.Brand) > 100 {
		response.BadRequest(c, "品牌名称不能超过100个字符")
		return
	}
	// 不可变字段以库中为准
	goods := payload.Goods
	goods.ID = existing.ID
	goods.TenantID = existing.TenantID
	goods.CreatedAt = existing.CreatedAt
	units, prices, stocks := h.buildChildren(&goods, payload.PriceRowsRaw, payload.StockRowsRaw)
	if err := h.repo.UpdateWithChildren(ctx, &goods, units, prices, stocks); err != nil {
		response.ServerError(c, "更新货品失败")
		return
	}
	h.respondWithDetail(c, goods.ID, goods)
}

func (h *GoodsHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(ctx, id); err != nil {
		response.ServerError(c, "删除货品失败")
		return
	}
	response.OKMsg(c, "删除成功")
}

func (h *GoodsHandler) respondWithDetail(c *gin.Context, id int64, fallback model.Goods) {
	full, err := h.repo.GetByIDWithChildren(c.Request.Context(), id)
	if err != nil {
		response.OK(c, fallback)
		return
	}
	response.OK(c, full)
}

// buildChildren 将前端 JSON payload 转换为子表结构
func (h *GoodsHandler) buildChildren(g *model.Goods, priceRowsRaw, stockRowsRaw string) ([]model.GoodsUnit, []model.GoodsPrice, []model.GoodsStock) {
	// ── 主单位名（来自全局单位） ──
	mainUnit := ""
	if g.UnitID != nil {
		var u model.Unit
		if h.repo.DB.Where("id = ?", int64(*g.UnitID)).First(&u).Error == nil {
			mainUnit = u.Name
		}
	}

	// ── 单位：主单位(is_main=1) + 辅单位 ──
	units := []model.GoodsUnit{}
	for _, u := range g.Units {
		if strings.TrimSpace(u.UnitName) == "" {
			continue
		}
		units = append(units, model.GoodsUnit{
			UnitID: u.UnitID, UnitName: u.UnitName, Factor: u.Factor, IsMain: 0,
		})
	}
	if mainUnit != "" {
		units = append([]model.GoodsUnit{{UnitName: mainUnit, Factor: 1, IsMain: 1}}, units...)
	} else if len(units) > 0 {
		units[0].IsMain = 1
		mainUnit = units[0].UnitName
	}
	unitNames := make([]string, 0, len(units))
	for _, u := range units {
		unitNames = append(unitNames, u.UnitName)
	}
	if len(unitNames) == 0 {
		unitNames = []string{mainUnit}
	}

	// ── 价格明细 ──
	var pr map[string]map[string]map[string]interface{}
	if priceRowsRaw != "" {
		_ = json.Unmarshal([]byte(priceRowsRaw), &pr)
	}
	hasUnitKeys := false
	for k := range pr {
		if k != "__simple__" {
			hasUnitKeys = true
			break
		}
	}
	prices := []model.GoodsPrice{}
	if hasUnitKeys {
		for _, un := range unitNames {
			specs := pr[un]
			for _, sk := range sortedKeys(specs) {
				prices = append(prices, buildGoodsPrice(un, sk, specs[sk]))
			}
		}
	} else {
		row := map[string]interface{}{}
		if s, ok := pr["__simple__"]; ok {
			if r, ok := s[""]; ok {
				row = r
			}
		}
		if len(row) == 0 {
			// 无 price_rows 时用主表价格兜底
			row = map[string]interface{}{
				"purchase_price":  g.PurchasePrice,
				"retail_price":    g.RetailPrice,
				"wholesale_price": g.WholesalePrice,
			}
		}
		if mainUnit == "" && len(unitNames) > 0 {
			mainUnit = unitNames[0]
		}
		prices = append(prices, buildGoodsPrice(mainUnit, "", row))
	}

	// ── 库存（按门店/仓库：shopId -> specKey -> row） ──
	var sr map[string]map[string]map[string]interface{}
	if stockRowsRaw != "" {
		_ = json.Unmarshal([]byte(stockRowsRaw), &sr)
	}
	stocks := []model.GoodsStock{}
	for _, shopKey := range sortedKeys(sr) {
		shopID, _ := strconv.ParseInt(shopKey, 10, 64)
		specs := sr[shopKey]
		for _, sk := range sortedKeys(specs) {
			st := buildGoodsStock(sk, specs[sk])
			st.ShopID = shopID
			stocks = append(stocks, st)
		}
	}
	if len(stocks) == 0 {
		stocks = append(stocks, model.GoodsStock{
			SpecKey:  "",
			Stock:    g.CurrentStock,
			MinStock: g.MinStock,
			MaxStock: g.MaxStock,
			InitCost: g.InitCost,
		})
	}

	return units, prices, stocks
}

func buildGoodsPrice(unitKey, specKey string, row map[string]interface{}) model.GoodsPrice {
	p := model.GoodsPrice{UnitKey: unitKey, SpecKey: specKey, Custom: "{}"}
	if row == nil {
		return p
	}
	p.Code = strVal(row["code"])
	p.Barcode = strVal(row["barcode"])
	p.PurchasePrice = numVal(row["purchase_price"])
	p.RetailPrice = numVal(row["retail_price"])
	p.WholesalePrice = numVal(row["wholesale_price"])
	if b, ok := row["disabled"].(bool); ok && b {
		p.Disabled = 1
	}
	if cm, ok := row["custom"].(map[string]interface{}); ok && len(cm) > 0 {
		norm := make(map[string]float64, len(cm))
		for k, v := range cm {
			norm[k] = numVal(v)
		}
		raw, _ := json.Marshal(norm)
		p.Custom = string(raw)
	}
	return p
}

func buildGoodsStock(specKey string, row map[string]interface{}) model.GoodsStock {
	s := model.GoodsStock{SpecKey: specKey}
	if row == nil {
		return s
	}
	s.Stock = intVal(row["stock"])
	s.MinStock = intVal(row["min_stock"])
	s.SafetyStock = intVal(row["safe_stock"])
	s.MaxStock = intVal(row["max_stock"])
	s.InitCost = numVal(row["init_cost"])
	return s
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func numVal(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
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

func intVal(v interface{}) int {
	return int(numVal(v))
}

func strVal(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
