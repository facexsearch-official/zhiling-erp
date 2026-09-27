package handler

import (
	"strconv"
	"strings"

	"pisa_server/internal/model"
	"pisa_server/internal/pkg/context"
	"pisa_server/internal/pkg/response"
	"pisa_server/internal/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/* ═══════════ 组装单 ═══════════ */

type AssemblyHandler struct{ db *gorm.DB }

func NewAssemblyHandler(db *gorm.DB) *AssemblyHandler { return &AssemblyHandler{db: db} }

type assemblyItemReq struct {
	Kind     int8    `json:"kind"`
	GoodsID  int64   `json:"goods_id"`
	Quantity float64 `json:"quantity"`
	UnitCost float64 `json:"unit_cost"`
	Amount   float64 `json:"amount"`
	Remark   string  `json:"remark"`
}

type assemblyCreateReq struct {
	SalesmanID  int64             `json:"salesman_id"`
	BillDate    string            `json:"bill_date"`
	AssemblyFee float64           `json:"assembly_fee"`
	AccountID   int64             `json:"account_id"`
	Remark      string            `json:"remark"`
	Items       []assemblyItemReq `json:"items"`
	SaveRecipe  bool              `json:"save_recipe"`
}

func (h *AssemblyHandler) fillItems(items []model.AssemblyItem) []model.AssemblyItem {
	var ids []int64
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	gm := map[int64]model.Goods{}
	if len(ids) > 0 {
		var goods []model.Goods
		h.db.Where("id IN ?", ids).Find(&goods)
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
	return items
}

func (h *AssemblyHandler) List(c *gin.Context) { h.list(c, 1) }

func (h *AssemblyHandler) ListSplit(c *gin.Context) { h.list(c, 2) }

func (h *AssemblyHandler) list(c *gin.Context, atype int8) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	hideVoid := c.Query("hide_void") == "1"

	var total int64
	var list []model.Assembly
	q := h.db.Table("assemblies AS p").
		Joins("LEFT JOIN salesmen sm ON sm.id = p.salesman_id").
		Joins("LEFT JOIN users u ON u.id = p.created_by").
		Where("p.tenant_id = ? AND p.type = ?", tenantID, atype)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("p.order_no LIKE ? OR p.remark LIKE ?", kw, kw)
	}
	if dateFrom != "" {
		q = q.Where("p.bill_date >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("p.bill_date <= ?", dateTo)
	}
	if hideVoid {
		q = q.Where("p.status = 1")
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Select("p.*, sm.name AS salesman_name, u.nickname AS maker_name").
		Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)

	var comps []model.AssemblyItem
	if len(list) > 0 {
		var ids []int64
		for _, a := range list {
			ids = append(ids, a.ID)
		}
		h.db.Table("assembly_items").Where("assembly_id IN ? AND kind = 1", ids).Order("id ASC").Find(&comps)
		comps = h.fillItems(comps)
	}
	cm := map[int64]model.AssemblyItem{}
	for _, ci := range comps {
		if _, ok := cm[ci.AssemblyID]; !ok {
			cm[ci.AssemblyID] = ci
		}
	}
	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
		if ci, ok := cm[list[i].ID]; ok {
			list[i].ComponentName = ci.GoodsName
			list[i].ComponentQty = ci.Quantity
		}
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *AssemblyHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var a model.Assembly
	if err := h.db.Table("assemblies").Where("id = ? AND tenant_id = ?", id, tenantID).First(&a).Error; err != nil {
		response.NotFound(c, "组装单不存在")
		return
	}
	a.IDStr = strconv.FormatInt(a.ID, 10)
	if a.SalesmanID != 0 {
		var sm model.Salesman
		if h.db.Where("id = ?", a.SalesmanID).First(&sm).Error == nil {
			a.SalesmanName = sm.Name
		}
	}
	if a.AccountID != 0 {
		var ac model.Account
		if h.db.Where("id = ?", a.AccountID).First(&ac).Error == nil {
			a.AccountName = ac.Name
		}
	}
	if a.CreatedBy != 0 {
		var u model.User
		if h.db.Where("id = ?", a.CreatedBy).First(&u).Error == nil {
			a.MakerName = u.Nickname
		}
	}
	var items []model.AssemblyItem
	h.db.Table("assembly_items").Where("assembly_id = ?", a.ID).Order("id ASC").Find(&items)
	a.Items = h.fillItems(items)
	response.OK(c, a)
}

func (h *AssemblyHandler) buildItems(tenantID, assemblyID int64, req []assemblyItemReq) []model.AssemblyItem {
	items := make([]model.AssemblyItem, 0, len(req))
	for _, it := range req {
		if it.GoodsID == 0 {
			continue
		}
		kind := it.Kind
		if kind == 0 {
			kind = 2
		}
		items = append(items, model.AssemblyItem{
			ID: snowflake.GenID(), TenantID: tenantID, AssemblyID: assemblyID, Kind: kind,
			GoodsID: it.GoodsID, Quantity: it.Quantity, UnitCost: it.UnitCost,
			Amount: round2o(it.Quantity * it.UnitCost), Remark: it.Remark,
		})
	}
	return items
}

func (h *AssemblyHandler) Create(c *gin.Context) { h.create(c, 1, "ZHD") }

func (h *AssemblyHandler) CreateSplit(c *gin.Context) { h.create(c, 2, "CFD") }

func (h *AssemblyHandler) create(c *gin.Context, atype int8, prefix string) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	userID := context.GetUserID(ctx)
	var req assemblyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if len(req.Items) == 0 {
		response.BadRequest(c, "请添加明细")
		return
	}
	a := model.Assembly{
		ID: snowflake.GenID(), TenantID: tenantID, Type: atype,
		OrderNo: nextNo(h.db, tenantID, "assemblies", prefix), SalesmanID: req.SalesmanID,
		BillDate: req.BillDate, AssemblyFee: req.AssemblyFee, AccountID: req.AccountID,
		Status: 1, Remark: req.Remark, CreatedBy: userID,
	}
	if err := h.db.Table("assemblies").Create(&a).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	items := h.buildItems(tenantID, a.ID, req.Items)
	if len(items) > 0 {
		h.db.Table("assembly_items").Create(&items)
	}
	if req.SaveRecipe {
		h.saveRecipeFromItems(tenantID, items, req.Remark)
	}
	a.IDStr = strconv.FormatInt(a.ID, 10)
	a.Items = h.fillItems(items)
	response.OK(c, a)
}

func (h *AssemblyHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req assemblyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var a model.Assembly
	if err := h.db.Table("assemblies").Where("id = ? AND tenant_id = ?", id, tenantID).First(&a).Error; err != nil {
		response.NotFound(c, "组装单不存在")
		return
	}
	h.db.Table("assemblies").Where("id = ?", id).Updates(map[string]interface{}{
		"salesman_id": req.SalesmanID, "bill_date": req.BillDate, "assembly_fee": req.AssemblyFee,
		"account_id": req.AccountID, "remark": req.Remark,
	})
	h.db.Table("assembly_items").Where("assembly_id = ?", id).Delete(&model.AssemblyItem{})
	items := h.buildItems(tenantID, id, req.Items)
	if len(items) > 0 {
		h.db.Table("assembly_items").Create(&items)
	}
	a.IDStr = strconv.FormatInt(a.ID, 10)
	a.Items = h.fillItems(items)
	response.OK(c, a)
}

func (h *AssemblyHandler) Void(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("assemblies").Where("id = ? AND tenant_id = ?", id, tenantID).Update("status", 9)
	response.OKMsg(c, "作废成功")
}

func (h *AssemblyHandler) saveRecipeFromItems(tenantID int64, items []model.AssemblyItem, remark string) {
	var component *model.AssemblyItem
	subs := []model.AssemblyItem{}
	for i := range items {
		if items[i].Kind == 1 && component == nil {
			component = &items[i]
		} else if items[i].Kind == 2 {
			subs = append(subs, items[i])
		}
	}
	if component == nil || len(subs) == 0 {
		return
	}
	r := model.Recipe{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: component.GoodsID, Quantity: component.Quantity, Remark: remark}
	h.db.Table("recipes").Create(&r)
	ris := make([]model.RecipeItem, 0, len(subs))
	for _, s := range subs {
		ris = append(ris, model.RecipeItem{ID: snowflake.GenID(), TenantID: tenantID, RecipeID: r.ID, GoodsID: s.GoodsID, Quantity: s.Quantity, UnitCost: s.UnitCost, Remark: s.Remark})
	}
	h.db.Table("recipe_items").Create(&ris)
}

/* ═══════════ 配方 ═══════════ */

type RecipeHandler struct{ db *gorm.DB }

func NewRecipeHandler(db *gorm.DB) *RecipeHandler { return &RecipeHandler{db: db} }

type recipeItemReq struct {
	GoodsID  int64   `json:"goods_id"`
	Quantity float64 `json:"quantity"`
	UnitCost float64 `json:"unit_cost"`
	Remark   string  `json:"remark"`
}

type recipeCreateReq struct {
	GoodsID  int64           `json:"goods_id"`
	Quantity float64         `json:"quantity"`
	Remark   string          `json:"remark"`
	Items    []recipeItemReq `json:"items"`
}

func (h *RecipeHandler) fillItems(items []model.RecipeItem) []model.RecipeItem {
	var ids []int64
	for _, it := range items {
		ids = append(ids, it.GoodsID)
	}
	gm := map[int64]model.Goods{}
	if len(ids) > 0 {
		var goods []model.Goods
		h.db.Where("id IN ?", ids).Find(&goods)
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
	return items
}

func (h *RecipeHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	var total int64
	var list []model.Recipe
	q := h.db.Table("recipes AS p").Where("p.tenant_id = ?", tenantID)
	if keyword != "" {
		kw := "%" + keyword + "%"
		var gids []int64
		h.db.Model(&model.Goods{}).Where("tenant_id = ? AND (name LIKE ? OR code LIKE ? OR barcode LIKE ?)", tenantID, kw, kw, kw).Pluck("id", &gids)
		if len(gids) > 0 {
			q = q.Where("p.goods_id IN ?", gids)
		} else {
			q = q.Where("1 = 0")
		}
	}
	q.Session(&gorm.Session{}).Count(&total)
	q.Offset((page - 1) * pageSize).Limit(pageSize).Order("p.created_at DESC").Scan(&list)

	for i := range list {
		list[i].IDStr = strconv.FormatInt(list[i].ID, 10)
		var g model.Goods
		if h.db.Where("id = ?", list[i].GoodsID).First(&g).Error == nil {
			list[i].GoodsName = g.Name
			list[i].GoodsCode = g.Code
			list[i].UnitName = g.MainUnit
			list[i].Spec = orderSpecNames(g.SpecGroups)
			list[i].Barcode = g.Barcode
			list[i].ImageURL = g.ImageURL
		}
		var items []model.RecipeItem
		h.db.Table("recipe_items").Where("recipe_id = ?", list[i].ID).Order("id ASC").Find(&items)
		items = h.fillItems(items)
		parts := make([]string, 0, len(items))
		for _, it := range items {
			nm := it.GoodsName
			if it.Spec != "" {
				nm += " - " + it.Spec
			}
			parts = append(parts, nm+" x "+formatQty(it.Quantity)+it.UnitName)
		}
		list[i].SubSummary = strings.Join(parts, "、")
	}
	response.OKPage(c, list, total, page, pageSize)
}

func (h *RecipeHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r model.Recipe
	if err := h.db.Table("recipes").Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "配方不存在")
		return
	}
	r.IDStr = strconv.FormatInt(r.ID, 10)
	var g model.Goods
	if h.db.Where("id = ?", r.GoodsID).First(&g).Error == nil {
		r.GoodsName = g.Name
		r.GoodsCode = g.Code
		r.UnitName = g.MainUnit
		r.Spec = orderSpecNames(g.SpecGroups)
		r.Barcode = g.Barcode
		r.ImageURL = g.ImageURL
	}
	var items []model.RecipeItem
	h.db.Table("recipe_items").Where("recipe_id = ?", r.ID).Order("id ASC").Find(&items)
	r.Items = h.fillItems(items)
	response.OK(c, r)
}

func (h *RecipeHandler) buildItems(tenantID, recipeID int64, req []recipeItemReq) []model.RecipeItem {
	items := make([]model.RecipeItem, 0, len(req))
	for _, it := range req {
		if it.GoodsID == 0 {
			continue
		}
		items = append(items, model.RecipeItem{
			ID: snowflake.GenID(), TenantID: tenantID, RecipeID: recipeID,
			GoodsID: it.GoodsID, Quantity: it.Quantity, UnitCost: it.UnitCost, Remark: it.Remark,
		})
	}
	return items
}

func (h *RecipeHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	var req recipeCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.GoodsID == 0 {
		response.BadRequest(c, "请选择组合件商品")
		return
	}
	r := model.Recipe{ID: snowflake.GenID(), TenantID: tenantID, GoodsID: req.GoodsID, Quantity: req.Quantity, Remark: req.Remark}
	if err := h.db.Table("recipes").Create(&r).Error; err != nil {
		response.ServerError(c, err.Error())
		return
	}
	items := h.buildItems(tenantID, r.ID, req.Items)
	if len(items) > 0 {
		h.db.Table("recipe_items").Create(&items)
	}
	r.IDStr = strconv.FormatInt(r.ID, 10)
	r.Items = h.fillItems(items)
	response.OK(c, r)
}

func (h *RecipeHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req recipeCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	var r model.Recipe
	if err := h.db.Table("recipes").Where("id = ? AND tenant_id = ?", id, tenantID).First(&r).Error; err != nil {
		response.NotFound(c, "配方不存在")
		return
	}
	h.db.Table("recipes").Where("id = ?", id).Updates(map[string]interface{}{
		"goods_id": req.GoodsID, "quantity": req.Quantity, "remark": req.Remark,
	})
	h.db.Table("recipe_items").Where("recipe_id = ?", id).Delete(&model.RecipeItem{})
	items := h.buildItems(tenantID, id, req.Items)
	if len(items) > 0 {
		h.db.Table("recipe_items").Create(&items)
	}
	r.IDStr = strconv.FormatInt(r.ID, 10)
	r.Items = h.fillItems(items)
	response.OK(c, r)
}

func (h *RecipeHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := context.GetTenantID(ctx)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.db.Table("recipe_items").Where("recipe_id = ?", id).Delete(&model.RecipeItem{})
	h.db.Table("recipes").Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Recipe{})
	response.OKMsg(c, "删除成功")
}

func formatQty(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" {
		s = "0"
	}
	return s
}
