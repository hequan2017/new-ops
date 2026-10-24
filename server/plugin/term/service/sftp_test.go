package service

import "testing"

// TestNormalizeSftpPath 路径规范化不变量：结果恒为 "/" 开头的绝对路径，
// 相对输入锚定根、.. 越级被 Clean 吸收（SFTP 语义=授予所选主机文件面，不允许越出 /）。
func TestNormalizeSftpPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/root"},
		{"   ", "/root"}, // 空白输入等价空串（TrimSpace 后走默认）
		{"var/log", "/var/log"},
		{"/var/log", "/var/log"},
		{"../../etc/passwd", "/etc/passwd"}, // 越级被吸收
		{"/a/b/../../c", "/c"},
		{"a//b///c", "/a/b/c"},
	}
	for _, c := range cases {
		if got := normalizeSftpPath(c.in); got != c.want {
			t.Fatalf("normalizeSftpPath(%q)=%q want %q", c.in, got, c.want)
		}
	}
}
