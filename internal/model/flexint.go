package model

import (
	"strconv"
	"strings"
)

// FlexInt64 可同时从 JSON 数字或字符串解析的 int64，用于雪花 ID 精度安全传递。
type FlexInt64 int64

// UnmarshalJSON 支持数字 / 字符串 / null
func (f *FlexInt64) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	s = strings.Trim(s, `"`)
	if s == "" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*f = FlexInt64(v)
	return nil
}

// MarshalJSON 输出为数字
func (f FlexInt64) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(f), 10)), nil
}
