package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ComboHandler struct {
	db *gorm.DB
}

func NewComboHandler(dbConn *gorm.DB) *ComboHandler {
	return &ComboHandler{db: dbConn}
}

type comboItemReq struct {
	GoodsID        int64   `json:"goods_id"`
	Quantity       int     `json:"quantity"`
	PurchasePrice  float64 `json:"purchase_price"`
	RetailPrice    float64 `json:"retail_price"`
	WholesalePrice float64 `json:"wholesale_price"`
	Remark         string  `json:"remark"`
}

type comboReq struct {
	Name           string         `json:"name"`
	Barcode        string         `json:"barcode"`
	UnitID         *int64         `json:"unit_id"`
	UnitName       string         `json:"unit_name"`
	CategoryID     *int64         `json:"category_id"`
	ImageURL       string         `json:"image_url"`
	RetailPrice    float64        `json:"retail_price"`
	WholesalePrice float64        `json:"wholesale_price"`
	PriceColumns   string         `json:"price_columns"`
	CustomPrices   string         `json:"custom_prices"`
	Remark         string         `json:"remark"`
	Items          []comboItemReq `json:"items"`
}

func (h *ComboHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	var total int64
	var list []model.Combo
	q := h.db.Table("combos AS c").
		Joins("LEFT JOIN goods_categories g ON g.id = c.category_id").
		Where("c.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("c.name LIKE ? OR c.barcode LIKE ? OR c.remark LIKE ?", kw, kw, kw)
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("c.*, g.name AS category_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("c.created_at DESC").Scan(&list)

	// 套餐组成摘要
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
		var items []model.ComboItem
		h.db.Table("combo_items").Where("combo_id = ?", list[i].ID).Find(&items)
		parts := make([]string, 0, len(items))
		for _, it := range items {
			var g model.Goods
			name := ""
			spec := ""
			if h.db.Where("id = ?", it.GoodsID).First(&g).Error == nil {
				name = g.Name
				spec = orderSpecNames(g.SpecGroups)
			}
			s := name
			if spec != "" {
				s += "-" + spec
			}
			if it.Quantity > 0 {
				s += "*" + strconv.Itoa(it.Quantity)
			}
			if it.UnitName != "" {
				s += it.UnitName
			}
			parts = append(parts, s)
		}
		list[i].Summary = strings.Join(parts, ",")
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *ComboHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var combo model.Combo
	if err := h.db.Table("combos").Where("id = ? AND tenant_id = ?", id, tenantID).First(&combo).Error; err != nil {
		response.NotFound(c, "套餐不存在")
		return
	}
	combo.IDStr = strconv.FormatInt(combo.ID, 10)
	var items []model.ComboItem
	h.db.Table("combo_items").Where("combo_id = ?", combo.ID).Order("id ASC").Find(&items)
	var ids []int64
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	gm := map[int64]model.Goods{}
	if len(ids) > 0 {
		var goods []model.Goods
		h.db.Where("tenant_id = ? AND id IN ?", tenantID, ids).Find(&goods)
		for _, g := range goods {
			gm[g.ID] = g
		}
	}
	for i := range items {
		g := gm[items[i].GoodsID]
		items[i].GoodsName = g.Name
		items[i].GoodsCode = g.Code
		items[i].UnitName = g.MainUnit
		items[i].Spec = orderSpecNames(g.SpecGroups)
		items[i].Barcode = g.Barcode
		items[i].ImageURL = g.ImageURL
	}
	combo.Items = items
	response.OK(c, combo)
}

func (h *ComboHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req comboReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "请输入套餐名称")
		return
	}

	var tp, tr, tw float64
	items := make([]model.ComboItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.GoodsID == 0 {
			continue
		}
		tp += it.PurchasePrice * float64(it.Quantity)
		tr += it.RetailPrice * float64(it.Quantity)
		tw += it.WholesalePrice * float64(it.Quantity)
		items = append(items, model.ComboItem{
			ID: snowflake.GenID(), TenantID: tenantID, GoodsID: it.GoodsID, Quantity: it.Quantity,
			PurchasePrice: it.PurchasePrice, RetailPrice: it.RetailPrice, WholesalePrice: it.WholesalePrice, Remark: it.Remark,
		})
	}
	combo := model.Combo{
		ID: snowflake.GenID(), TenantID: tenantID, Name: req.Name, Barcode: req.Barcode,
		UnitID: req.UnitID, UnitName: req.UnitName, CategoryID: req.CategoryID, ImageURL: req.ImageURL,
		TotalPurchase: round2o(tp), TotalRetail: round2o(tr), TotalWholesale: round2o(tw),
		RetailPrice: req.RetailPrice, WholesalePrice: req.WholesalePrice,
		PriceColumns: req.PriceColumns, CustomPrices: req.CustomPrices,
		Status: 1, Remark: req.Remark,
	}
	if err := h.db.Table("combos").Create(&combo).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range items {
		items[i].ComboID = combo.ID
	}
	if len(items) > 0 {
		h.db.Table("combo_items").Create(&items)
	}
	combo.Items = items
	combo.IDStr = strconv.FormatInt(combo.ID, 10)
	response.OK(c, combo)
}

func (h *ComboHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("combo_items").Where("combo_id = ?", id).Delete(&model.ComboItem{})
	h.db.Table("combos").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Combo{})
	response.OKMsg(c, "删除成功")
}

// PriceLevels 返回全站价格等级名称（取自货品 price_columns 的并集）
func (h *ComboHandler) PriceLevels(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var cols []string
	h.db.Model(&model.Goods{}).Where("tenant_id = ? AND price_columns != ''", tenantID).Pluck("price_columns", &cols)
	seen := map[string]bool{}
	out := []string{}
	for _, c := range cols {
		var names []string
		if json.Unmarshal([]byte(c), &names) != nil {
			continue
		}
		for _, n := range names {
			if n != "" && !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	response.OK(c, out)
}
