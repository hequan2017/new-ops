package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hequan2017/new-ops/server/global"
	assetModel "github.com/hequan2017/new-ops/server/plugin/asset/model"
)

func hostOf(id uint, ip string, jump *uint) *assetModel.AssetHost {
	return &assetModel.AssetHost{
		GVA_MODEL:  global.GVA_MODEL{ID: id},
		Hostname:   fmt.Sprintf("h%d", id),
		IP:         ip,
		JumpHostID: jump,
	}
}

// lookupFrom 从 map 取主机（单测用 lookup 注入）
func lookupFrom(m map[uint]*assetModel.AssetHost) func(uint) (*assetModel.AssetHost, error) {
	return func(id uint) (*assetModel.AssetHost, error) {
		if h, ok := m[id]; ok {
			return h, nil
		}
		return nil, errors.New("record not found")
	}
}

func TestBuildJumpChain_Direct(t *testing.T) {
	target := hostOf(1, "10.0.0.1", nil)
	chain, err := buildJumpChain(target, lookupFrom(nil))
	if err != nil {
		t.Fatalf("直连不应报错: %v", err)
	}
	if len(chain) != 1 || chain[0] != target {
		t.Fatalf("直连链应仅含目标: %v", chain)
	}
}

func TestBuildJumpChain_TwoHops(t *testing.T) {
	// target(1) → jump(2) → jump(3)：链应为 [1,2,3]
	h3 := hostOf(3, "10.0.0.3", nil)
	h2 := hostOf(2, "10.0.0.2", u32(3))
	h1 := hostOf(1, "10.0.0.1", u32(2))
	chain, err := buildJumpChain(h1, lookupFrom(map[uint]*assetModel.AssetHost{2: h2, 3: h3}))
	if err != nil {
		t.Fatalf("两级跳板不应报错: %v", err)
	}
	if len(chain) != 3 || chain[0].ID != 1 || chain[1].ID != 2 || chain[2].ID != 3 {
		t.Fatalf("链序应为 [目标,跳板2,跳板3]: %v", chainIDs(chain))
	}
}

func TestBuildJumpChain_Loop(t *testing.T) {
	// 1 → 2 → 1 成环（lookup 含目标自身，等价 DB 场景）
	h2 := hostOf(2, "10.0.0.2", u32(1))
	h1 := hostOf(1, "10.0.0.1", u32(2))
	if _, err := buildJumpChain(h1, lookupFrom(map[uint]*assetModel.AssetHost{1: h1, 2: h2})); err == nil {
		t.Fatal("成环链必须报错")
	} else {
		te, ok := err.(*TermError)
		if !ok || te.Code != ErrCodeJumpLoop {
			t.Fatalf("应返回环错误码 %d: %v", ErrCodeJumpLoop, err)
		}
	}
	// 自环
	self := hostOf(9, "10.0.0.9", u32(9))
	if _, err := buildJumpChain(self, lookupFrom(map[uint]*assetModel.AssetHost{9: self})); err == nil {
		t.Fatal("自环必须报错")
	}
}

func TestBuildJumpChain_TooDeep(t *testing.T) {
	// 1→2→3→4→5→6→7：跳板 6 层（不含目标）超上限
	var nilPtr *uint
	m := map[uint]*assetModel.AssetHost{
		7: hostOf(7, "10.0.0.7", nilPtr),
		6: hostOf(6, "10.0.0.6", u32(7)),
		5: hostOf(5, "10.0.0.5", u32(6)),
		4: hostOf(4, "10.0.0.4", u32(5)),
		3: hostOf(3, "10.0.0.3", u32(4)),
		2: hostOf(2, "10.0.0.2", u32(3)),
		1: hostOf(1, "10.0.0.1", u32(2)),
	}
	_, err := buildJumpChain(m[1], lookupFrom(m))
	if err == nil {
		t.Fatal("超过 5 层跳板必须报错")
	}
	te, ok := err.(*TermError)
	if !ok || te.Code != ErrCodeJumpTooDeep {
		t.Fatalf("应返回超深错误码 %d: %v", ErrCodeJumpTooDeep, err)
	}
	// 恰好 5 层跳板应通过：1→2→3→4→5→6（6 为最外层，无上级）
	h6 := hostOf(6, "10.0.0.6", nilPtr)
	h5 := hostOf(5, "10.0.0.5", u32(6))
	h4 := hostOf(4, "10.0.0.4", u32(5))
	h3 := hostOf(3, "10.0.0.3", u32(4))
	h2 := hostOf(2, "10.0.0.2", u32(3))
	h1 := hostOf(1, "10.0.0.1", u32(2))
	chain, err := buildJumpChain(h1, lookupFrom(map[uint]*assetModel.AssetHost{2: h2, 3: h3, 4: h4, 5: h5, 6: h6}))
	if err != nil {
		t.Fatalf("恰好 5 层跳板应放行: %v", err)
	}
	if len(chain) != 6 {
		t.Fatalf("链长应为 6（目标+5跳）: %d", len(chain))
	}
}

func TestBuildJumpChain_JumpMissing(t *testing.T) {
	h1 := hostOf(1, "10.0.0.1", u32(99))
	_, err := buildJumpChain(h1, lookupFrom(nil))
	if err == nil || err.Error() == "" {
		t.Fatal("跳板不存在应报错")
	}
}

func TestBuildJumpChain_JumpNoIP(t *testing.T) {
	h2 := hostOf(2, "", nil)
	h1 := hostOf(1, "10.0.0.1", u32(2))
	if _, err := buildJumpChain(h1, lookupFrom(map[uint]*assetModel.AssetHost{2: h2})); err == nil {
		t.Fatal("跳板缺内网IP应报错")
	}
}

// ---------- 测试工具 ----------

func u32(v uint) *uint { return &v }

func chainIDs(chain []*assetModel.AssetHost) []uint {
	ids := make([]uint, 0, len(chain))
	for _, h := range chain {
		ids = append(ids, h.ID)
	}
	return ids
}
