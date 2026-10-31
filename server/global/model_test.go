package global

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGvaModelTimeSerialization(t *testing.T) {
	ts := time.Date(2026, 10, 31, 15, 0, 0, 0, time.FixedZone("CST", 8*3600))
	m := GVA_MODEL{ID: 1, CreatedAt: ts, UpdatedAt: ts}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	// 前端契约：字段名 CreatedAt/UpdatedAt（大写开头，多处 vue 模板消费）；值 RFC3339 带时区可被 new Date 解析
	if _, ok := out["CreatedAt"]; !ok {
		t.Fatalf("CreatedAt 键缺失: %s", b)
	}
	s, _ := out["CreatedAt"].(string)
	if _, err := time.Parse(time.RFC3339, s); err != nil {
		t.Fatalf("CreatedAt 非可解析的 RFC3339: %s", s)
	}
	if _, ok := out["DeletedAt"]; ok {
		t.Fatal("DeletedAt 应 json:\"-\" 不序列化")
	}
}
