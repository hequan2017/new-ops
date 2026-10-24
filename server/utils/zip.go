package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// unzipTotalSizeLimit 解压总大小上限（解压炸弹防护）
const unzipTotalSizeLimit = 512 << 20

// 解压
func Unzip(zipFile string, destDir string) ([]string, error) {
	zipReader, err := zip.OpenReader(zipFile)
	var paths []string
	if err != nil {
		return []string{}, err
	}
	defer zipReader.Close()

	var total uint64
	for _, f := range zipReader.File {
		if strings.Contains(f.Name, "..") {
			return []string{}, fmt.Errorf("%s 文件名不合法", f.Name)
		}
		fpath := filepath.Join(destDir, f.Name)
		paths = append(paths, fpath)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return []string{}, err
		}
		if err = extractZipEntry(f, fpath, &total); err != nil {
			return []string{}, err
		}
	}
	return paths, nil
}

// extractZipEntry 写出单个条目：句柄在函数作用域内闭合（原实现在循环里 defer，条目多时句柄累积），
// 并对累计解压体积设上限，超限即中断（LimitReader 截断在写入中途生效，不先落完整个大文件）
func extractZipEntry(f *zip.File, fpath string, total *uint64) error {
	inFile, err := f.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	remaining := int64(unzipTotalSizeLimit) - int64(*total)
	n, err := io.Copy(outFile, io.LimitReader(inFile, remaining+1))
	if err != nil {
		return err
	}
	*total += uint64(n)
	if *total > unzipTotalSizeLimit {
		return fmt.Errorf("解压总大小超过 %d MB 上限", unzipTotalSizeLimit>>20)
	}
	return nil
}
