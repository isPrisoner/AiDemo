package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var APIKey string

func LoadEnv() error {
	// 尝试加载init/initApi.env文件
	err := godotenv.Load("init/initApi.env")
	if err != nil {
		return fmt.Errorf("加载.env文件失败: %w", err)
	}

	APIKey = os.Getenv("DOUBAO_API_KEY")
	if APIKey == "" {
		return fmt.Errorf("请在.env文件中设置 DOUBAO_API_KEY")
	}

	return nil
}
