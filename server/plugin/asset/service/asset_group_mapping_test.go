package service

import "testing"

func TestGroupHostIDsByGroup(t *testing.T) {
	rows := []groupMappingRow{
		{GroupID: 2, HostID: 22},
		{GroupID: 1, HostID: 11},
		{GroupID: 2, HostID: 23},
		{GroupID: 3, HostID: 31},
	}
	got := groupHostIDsByGroup(rows)
	if len(got[2]) != 2 || got[2][0] != 22 || got[2][1] != 23 {
		t.Fatalf("组2 主机聚合不符: %v", got[2])
	}
	if len(got[1]) != 1 || got[1][0] != 11 {
		t.Fatalf("组1 主机聚合不符: %v", got[1])
	}
	if _, ok := got[9]; ok {
		t.Fatal("不存在的组不应出现")
	}
	// 空输入 → 空表
	if got := groupHostIDsByGroup(nil); len(got) != 0 {
		t.Fatalf("空输入应返回空表: %v", got)
	}
}

func TestGroupUserIDsByGroup(t *testing.T) {
	rows := []groupMappingRow{
		{GroupID: 1, UserID: 7},
		{GroupID: 1, UserID: 8},
		{GroupID: 2, UserID: 9},
	}
	got := groupUserIDsByGroup(rows)
	if len(got[1]) != 2 || got[1][0] != 7 || got[1][1] != 8 {
		t.Fatalf("组1 用户聚合不符: %v", got[1])
	}
	if len(got[2]) != 1 || got[2][0] != 9 {
		t.Fatalf("组2 用户聚合不符: %v", got[2])
	}
}
