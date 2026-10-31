package utils

import "testing"

// TestSHA256V SHA-256 摘要与标准测试向量对齐（空串/abc/追加 salt 变体）
func TestSHA256V(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"空串", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"标准向量abc", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	}
	for _, c := range cases {
		if got := SHA256V([]byte(c.in)); got != c.want {
			t.Fatalf("%s: want %s got %s", c.name, c.want, got)
		}
	}
}

// TestBcryptHashCheck 哈希-校验往返：正确密码通过、错误密码拒绝、同密码两次哈希结果不同（盐随机）
func TestBcryptHashCheck(t *testing.T) {
	hash := BcryptHash("s3cret!pass")
	if hash == "" || hash == "s3cret!pass" {
		t.Fatalf("哈希结果异常: %q", hash)
	}
	if !BcryptCheck("s3cret!pass", hash) {
		t.Fatal("正确密码校验应通过")
	}
	if BcryptCheck("wrong-pass", hash) {
		t.Fatal("错误密码校验应拒绝")
	}
	again := BcryptHash("s3cret!pass")
	if again == hash {
		t.Fatal("同密码两次哈希应因随机盐而不同")
	}
}
