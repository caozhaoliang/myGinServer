package userserver

import (
	"regexp"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// 密码加密
func hashPassword(password string) (string, error) {
	// 使用bcrypt算法加密密码，Cost值越高加密越慢但越安全
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// 验证密码复杂度
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度不能少于8个字符")
	}

	// 检查是否包含数字
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("密码必须包含至少一个数字")
	}

	// 检查是否包含大写字母
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("密码必须包含至少一个大写字母")
	}

	// 检查是否包含小写字母
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("密码必须包含至少一个小写字母")
	}

	// 检查是否包含特殊字符
	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password) {
		return errors.New("密码必须包含至少一个特殊字符")
	}

	return nil
}
