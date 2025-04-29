package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	// 定义命令行参数：端口号和目录
	port := flag.Int("port", 8080, "HTTP server port")
	dir := flag.String("dir", ".", "Directory to serve files from")
	flag.Parse()

	// 创建 Gin 引擎
	r := gin.Default()

	// 将指定目录作为静态文件根目录，并启用目录列表展示
	r.StaticFS("/", gin.Dir(*dir, true))

	// 启动服务
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Serving %s on HTTP port: %d with directory listing enabled\n", *dir, *port)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
