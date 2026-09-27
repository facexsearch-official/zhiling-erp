package model

import "time"

// CommissionRule 提成规则
type CommissionRule struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	IDStr         string    `json:"id_str" gorm:"->"`
	TenantID      int64     `json:"tenant_id" gorm:"index"`
	Name          string    `json:"name" gorm:"size:64"`
	Method        int8      `json:"method"`         // 1=按销售应收 2=按销售实收 3=按销售毛利
	DeductFreight int8      `json:"deduct_freight"` // 0=不扣除 1=扣除所含运费
	CalcType      int8      `json:"calc_type"`      // 1=统一提成 2=阶梯提成
	Ratio         float64   `json:"ratio"`          // 提成比例(%)
	Tiers         string    `json:"tiers" gorm:"type:text"`
	BonusEnabled  int8      `json:"bonus_enabled"`
	BonusTarget   float64   `json:"bonus_target"`
	BonusAmount   float64   `json:"bonus_amount"`
	StaffIDs      string    `json:"staff_ids" gorm:"type:text"`
	Status        int8      `json:"status" gorm:"default:1"`
	CreatedAt     time.Time `json:"created_at"`
}
