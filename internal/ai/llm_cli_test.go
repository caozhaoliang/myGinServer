package ai

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func Test(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	openaiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	modelName := "moonshot-v1-8k"
	// 初始化 OpenAI 模型
	llm, err := openai.New(
		openai.WithToken(openaiKey),
		openai.WithBaseURL(baseURL),
		openai.WithModel(modelName),
	)
	if err != nil {
		log.Fatalf("无法创建 OpenAI 客户端: %v", err)
	}

	// 创建提示词模板
	promptTemplate := prompts.NewPromptTemplate(
		"请用根据用户的评价: {{ .command }}输入用户的情绪值，其中1表示极好的，相当棒的。 2表示很不错，值得推荐的。3表示普通的，一般般。4表示不推荐，很乏味。5表示极差的，令人讨厌的；仅输出极简结论1-5",
		[]string{"command"},
	)

	// 填充模板参数
	prompt, err := promptTemplate.Format(map[string]interface{}{
		// "command": "我觉得<精舞门>这部电影整体看下来还是没什么尿点的，角色塑造还算成功，笑点也是一只在线，最后导演对剧情进一步做了升华。",
		"command": "<精舞门>真的非常非常好看，我和我女朋友都笑的合不拢嘴,真的血妈推荐大家来看",
	})
	if err != nil {
		log.Fatalf("格式化提示词失败: %v", err)
	}

	// 调用模型获取回答
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatalf("生成回答失败: %v", err)
	}

	// 输出结果
	fmt.Println("回答:", completion)
}
