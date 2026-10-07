package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

// BcryptHash 使用 bcrypt 对密码进行加密
func BcryptHash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

// BcryptCheck 对比明文密码和数据库的哈希值
func BcryptCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: SHA256V
//@description: SHA-256 摘要（原 MD5V 已升级：MD5 不满足安全要求，调用点均为文件命名/索引名/分片校验）
//@param: str []byte
//@return: string

func SHA256V(str []byte, b ...byte) string {
	h := sha256.New()
	h.Write(str)
	return hex.EncodeToString(h.Sum(b))
}
