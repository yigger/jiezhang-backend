package main

import (
	"log"

	"github.com/yigger/jiezhang-backend/internal/bootstrap"
)

// @title Jiezhang API
// @version 1.0
// @description 记账 API。部分业务错误使用 HTTP 200，通过 status/msg 判断结果。
// @BasePath /api
// @securityDefinitions.apikey AppID
// @in header
// @name X-WX-APP-ID
// @securityDefinitions.apikey SessionKey
// @in header
// @name X-WX-Skey
func main() {
	if err := bootstrap.Main(); err != nil {
		log.Fatal(err)
	}
}
