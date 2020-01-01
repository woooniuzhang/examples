package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"woniuxiaoan/examples/mutatingwebhook/business"
)

var (
	svcAddr = ":8088"
)

func main() {
	server := echo.New()
	addMiddleware(server)
	register(server)
	server.StartTLS(svcAddr, "./tls/tls.crt", "./tls/tls.key")
}

func register(server *echo.Echo) {
	server.POST("/mutate", business.MutateHandler)
}

func addMiddleware(server *echo.Echo) {
	server.Use(middleware.Recover())
}
