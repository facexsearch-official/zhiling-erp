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

type PriceLevelHandler struct{ db *gorm.DB }

func NewPriceLevelHandler(db *gorm.DB) *PriceLevelHandler { return &PriceLevelHandler{db: db} }

func (h *PriceLevelHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var cnt int64
	h.db.Model(&model.PriceLevel{}).Where("tenant_id = ?", tenantID).Count(&cnt)
	if cnt == 0 {
		defs := []model.PriceLevel{
			{ID: snowflake.GenID(), TenantID: tenantID, Name: "零售价", Sort: 0, Status: 1},
			{ID: snowflake.GenID(), TenantID: tenantID, Name: "批发价", Sort: 1, Status: 1},
		}
		h.db.Table("price_levels").Create(&defs)
	}
	var list []model.PriceLevel
	h.db.Where("tenant_id = ?", tenantID).Order("sort ASC, id ASC").Find(&list)
	out := make([]gin.H, 0, len(list))
	for _, x := range list {
		out = append(out, gin.H{
			"id":      x.ID,
			"id_str":  strconv.FormatInt(x.ID, 10),
			"name":    x.Name,
			"sort":    x.Sort,
			"status":  x.Status,
		})
	}
	response.OK(c, out)
}

type plReq struct {
	Name   string `json:"name"`
	Status int8   `json:"status"`
}

func (h *PriceLevelHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req plReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		response.BadRequest(c, "请输入价格等级名称")
		return
	}
	pl := model.PriceLevel{ID: snowflake.GenID(), TenantID: tenantID, Name: req.Name, Status: 1}
	h.db.Table("price_levels").Create(&pl)
	response.OK(c, pl)
}

func (h *PriceLevelHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req plReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	upd := map[string]interface{}{}
	if req.Name != "" {
		upd["name"] = req.Name
	}
	if req.Status != 0 {
		upd["status"] = req.Status
	}
	h.db.Table("price_levels").Where("id = ? AND tenant_id = ?", id, tenantID).Updates(upd)
	response.OKMsg(c, "更新成功")
}

func (h *PriceLevelHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("price_levels").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.PriceLevel{})
	response.OKMsg(c, "删除成功")
}
