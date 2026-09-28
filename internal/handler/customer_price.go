package handler

import (
	"strconv"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CustomerPriceHandler struct{ db *gorm.DB }

func NewCustomerPriceHandler(db *gorm.DB) *CustomerPriceHandler {
	return &CustomerPriceHandler{db: db}
}

func (h *CustomerPriceHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	q := h.db.Table("customer_prices AS cp").
		Joins("JOIN customers cu ON cu.id = cp.customer_id").
		Joins("JOIN goods g ON g.id = cp.goods_id").
		Where("cp.tenant_id = ?", tenantID)
	if v := c.Query("customer_id"); v != "" && v != "0" {
		q = q.Where("cp.customer_id = ?", v)
	}
	if v := c.Query("category_id"); v != "" && v != "0" {
		q = q.Where("g.category_id = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		kw := "%" + v + "%"
		q = q.Where("g.name LIKE ? OR g.barcode LIKE ? OR g.code LIKE ?", kw, kw, kw)
	}
	var total int64
	q.Session(&gorm.Session{}).Count(&total)

	type raw struct {
		ID           int64
		CustomerID   int64
		GoodsID      int64
		Price        float64
		CustomerName string
		PriceLevel   string
		GoodsName    string
		Code         string
		MainUnit     string
		SpecGroups   string
		ImageURL     string
	}
	var rows []raw
	q.Select("cp.id, cp.customer_id, cp.goods_id, cp.price, cu.name AS customer_name, cu.price_level, g.name AS goods_name, g.code, g.main_unit, g.spec_groups, g.image_url").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("cp.updated_at DESC").Scan(&rows)

	type outRow struct {
		IDStr        string  `json:"id_str"`
		CustomerID   int64   `json:"customer_id"`
		GoodsID      int64   `json:"goods_id"`
		Price        float64 `json:"price"`
		CustomerName string  `json:"customer_name"`
		PriceLevel   string  `json:"price_level"`
		GoodsName    string  `json:"goods_name"`
		Code         string  `json:"code"`
		MainUnit     string  `json:"main_unit"`
		Spec         string  `json:"spec"`
		ImageURL     string  `json:"image_url"`
	}
	list := make([]outRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, outRow{
			IDStr: strconv.FormatInt(r.ID, 10), CustomerID: r.CustomerID, GoodsID: r.GoodsID, Price: round2o(r.Price),
			CustomerName: r.CustomerName, PriceLevel: r.PriceLevel, GoodsName: r.GoodsName, Code: r.Code,
			MainUnit: r.MainUnit, Spec: orderSpecNames(r.SpecGroups), ImageURL: r.ImageURL,
		})
	}
	response.OKPage(c, list, total, page, pageSize)
}

// ByCustomer 根据客户报价：列出所有商品及该客户的最新报价
func (h *CustomerPriceHandler) ByCustomer(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	customerID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	type raw struct {
		GoodsID        int64
		Name           string
		Code           string
		MainUnit       string
		SpecGroups     string
		ImageURL       string
		RetailPrice    float64
		WholesalePrice float64
		Price          float64
	}
	var rows []raw
	h.db.Table("goods AS g").
		Joins("LEFT JOIN customer_prices cp ON cp.goods_id = g.id AND cp.customer_id = ?", customerID).
		Where("g.tenant_id = ?", tenantID).
		Select("g.id AS goods_id, g.name, g.code, g.main_unit, g.spec_groups, g.image_url, g.retail_price, g.wholesale_price, COALESCE(cp.price,0) AS price").
		Order("g.id DESC").Scan(&rows)

	type outRow struct {
		GoodsID        int64   `json:"goods_id"`
		IDStr          string  `json:"id_str"`
		Name           string  `json:"name"`
		Code           string  `json:"code"`
		Unit           string  `json:"unit"`
		Spec           string  `json:"spec"`
		ImageURL       string  `json:"image_url"`
		RetailPrice    float64 `json:"retail_price"`
		WholesalePrice float64 `json:"wholesale_price"`
		Price          float64 `json:"price"`
	}
	list := make([]outRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, outRow{
			GoodsID: r.GoodsID, IDStr: strconv.FormatInt(r.GoodsID, 10), Name: r.Name, Code: r.Code,
			Unit: r.MainUnit, Spec: orderSpecNames(r.SpecGroups), ImageURL: r.ImageURL,
			RetailPrice: r.RetailPrice, WholesalePrice: r.WholesalePrice, Price: round2o(r.Price),
		})
	}
	response.OK(c, list)
}

// ByGoods 根据商品报价：列出所有客户及该商品的最新报价
func (h *CustomerPriceHandler) ByGoods(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	goodsID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	type raw struct {
		CustomerID int64
		Name       string
		PriceLevel string
		Price      float64
	}
	var rows []raw
	h.db.Table("customers AS cu").
		Joins("LEFT JOIN customer_prices cp ON cp.customer_id = cu.id AND cp.goods_id = ?", goodsID).
		Where("cu.tenant_id = ? AND cu.status = 1", tenantID).
		Select("cu.id AS customer_id, cu.name, cu.price_level, COALESCE(cp.price,0) AS price").
		Order("cu.id DESC").Scan(&rows)

	type outRow struct {
		CustomerID int64   `json:"customer_id"`
		IDStr      string  `json:"id_str"`
		Name       string  `json:"name"`
		PriceLevel string  `json:"price_level"`
		Price      float64 `json:"price"`
	}
	list := make([]outRow, 0, len(rows))
	for _, r := range rows {
		list = append(list, outRow{CustomerID: r.CustomerID, IDStr: strconv.FormatInt(r.CustomerID, 10), Name: r.Name, PriceLevel: r.PriceLevel, Price: round2o(r.Price)})
	}
	response.OK(c, list)
}

type cpSaveItem struct {
	CustomerID int64   `json:"customer_id"`
	GoodsID    int64   `json:"goods_id"`
	Price      float64 `json:"price"`
}
type cpSaveReq struct {
	Items []cpSaveItem `json:"items"`
}

func (h *CustomerPriceHandler) Save(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req cpSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	for _, it := range req.Items {
		if it.CustomerID == 0 || it.GoodsID == 0 {
			continue
		}
		var existing model.CustomerPrice
		err := h.db.Where("tenant_id = ? AND customer_id = ? AND goods_id = ?", tenantID, it.CustomerID, it.GoodsID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			h.db.Table("customer_prices").Create(&model.CustomerPrice{
				ID: snowflake.GenID(), TenantID: tenantID, CustomerID: it.CustomerID, GoodsID: it.GoodsID, Price: it.Price,
			})
		} else if err == nil {
			h.db.Table("customer_prices").Where("id = ?", existing.ID).Update("price", it.Price)
		}
	}
	response.OKMsg(c, "保存成功")
}

func (h *CustomerPriceHandler) History(c *gin.Context) {
	response.OK(c, []interface{}{})
}
