package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

var APIKey string

func LoadEnv() {
	err := godotenv.Load("init/init.env")
	if err != nil {
		fmt.Println("加载.env文件失败:", err)
		os.Exit(1)
	}

	APIKey = os.Getenv("DOUBAO_API_KEY")
	if APIKey == "" {
		fmt.Println("请在.env文件中设置 DOUBAO_API_KEY")
		os.Exit(1)
	}
}
