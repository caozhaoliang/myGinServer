package utils

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestGenerateAvatars(t *testing.T) {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成5个不同的头像
	for i := 0; i < 1; i++ {
		width := 200
		height := 200

		outputPath := fmt.Sprintf("avatar_%d.svg", i+1)
		content := generateHumanAvatar(width, height)
		err := os.WriteFile(outputPath, []byte(content), 0644)
		fmt.Printf("成功生成头像: %s\n, %v", outputPath, err)
	}
}
