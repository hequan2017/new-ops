package utils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildTestZip 构造测试 zip（内容为 name→content 映射，自动含父目录条目）
func buildTestZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, content := range entries {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestUnzip 正常解压：多文件+子目录，内容逐字一致
func TestUnzip(t *testing.T) {
	src := buildTestZip(t, map[string]string{
		"a/b/hello.txt": "hello",
		"top.txt":       "world",
	})
	dest := t.TempDir()
	paths, err := Unzip(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 { // 标准库 Create 不生成隐式目录条目，仅 2 个文件条目（目录由 Unzip 的 MkdirAll 落盘）
		t.Fatalf("期望 2 个条目路径 got %d", len(paths))
	}
	for name, want := range map[string]string{"a/b/hello.txt": "hello", "top.txt": "world"} {
		got, err := os.ReadFile(filepath.Join(dest, name))
		if err != nil {
			t.Fatalf("解压产物缺失 %s: %v", name, err)
		}
		if string(got) != want {
			t.Fatalf("%s 内容不符 want %q got %q", name, want, string(got))
		}
	}
}

// TestUnzip_PathTraversal 拒绝含 .. 的条目名（zip 路径穿越防护）
func TestUnzip_PathTraversal(t *testing.T) {
	src := buildTestZip(t, map[string]string{"../evil.txt": "boom"})
	_, err := Unzip(src, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "文件名不合法") {
		t.Fatalf("应拒绝路径穿越, got %v", err)
	}
}
