package main

import (
	"log"
	"os"
)

func main() {
	// 创建一个日志文件
	logFile, err := os.Create("app.log")
	if err != nil {
		log.Fatal("Cannot create log file:", err)
	}
	defer logFile.Close()

	// 设置日志输出到文件
	log.SetOutput(logFile)

	// 示例日志记录
	log.Println("This is a log message.")
	log.Printf("Formatted log message: %s", "Hello, Log!")
	log.Fatal("a fatal log")
	// 记录错误日志
	log.Println("An error occurred:", err)
}
