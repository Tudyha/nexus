package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	minBackoff = 1 * time.Second
	maxBackoff = 60 * time.Second
)

// StringToUint64 字符串转uint64
func StringToUint64(str string) uint64 {
	r, _ := strconv.ParseUint(str, 10, 64)
	return r
}

// GenerateRandomNumber 生成随机数字
func GenerateRandomNumber(length int) string {
	const charset = "0123456789"
	result := make([]byte, length)
	for i := range result {
		idx := rand.IntN(len(charset))
		result[i] = charset[idx]
	}
	return string(result)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		idx := rand.IntN(len(charset))
		result[i] = charset[idx]
	}
	return string(result)
}

// Uint64ToString uint64转字符串
func Uint64ToString(num uint64) string {
	return strconv.FormatUint(num, 10)
}

// Base64Encode 将字节数组编码为Base64字符串
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// MD5 计算MD5
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func Checksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func VerifyChecksum(path, expected string) error {
	got, err := Checksum(path)
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("checksum mismatch: got %s, expected %s", got, expected)
	}
	return nil
}

// ValidatePassword 校验密码强度，返回错误描述。
// 规则：至少 8 位，包含大小写字母、数字、特殊字符。
func ValidatePassword(password string) error {
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)
	if len(password) < 8 {
		return fmt.Errorf("密码长度不能少于 8 位")
	}
	if len(password) > 128 {
		return fmt.Errorf("密码长度不能超过 128 位")
	}
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	var missing []string
	if !hasUpper {
		missing = append(missing, "大写字母")
	}
	if !hasLower {
		missing = append(missing, "小写字母")
	}
	if !hasDigit {
		missing = append(missing, "数字")
	}
	if !hasSpecial {
		missing = append(missing, "特殊字符")
	}
	if len(missing) > 0 {
		return fmt.Errorf("密码必须包含：%s", strings.Join(missing, "、"))
	}
	return nil
}

// 返回下一次重连等待时间，指数退避 + jitter。
func NextRetryDelay(retryAttempt int) time.Duration {
	shift := retryAttempt - 1
	if shift > 6 {
		shift = 6
	}
	// 2^(shift) seconds: 1s, 2s, 4s, 8s, 16s, 32s, 64s(max→60s)
	delay := minBackoff << shift
	if delay > maxBackoff {
		delay = maxBackoff
	}
	// ±25% jitter 避免 thundering herd
	jitter := time.Duration(float64(delay) * (rand.Float64()*0.5 - 0.25))
	delay += jitter
	if delay < minBackoff {
		delay = minBackoff
	}
	return delay
}
