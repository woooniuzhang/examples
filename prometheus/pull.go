package prometheus

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
)

//经典的pull模式, 适合long running 或者对监控实时性要求不是特别高的场景。

var pullCounterVec *prometheus.CounterVec

func init() {
	pullCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_total",
		Help: "request_total",
	}, []string{"method","path"})

	//将此metric信息注册进Gather中，这样在提供metrics接口时就可以
	//按照prometheus的方式将metric展现出来了
	prometheus.MustRegister(pullCounterVec)
}

func pullMiddleware(fn echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("middleware panic: %v", err)
			}
		}()

		pullCounterVec.With(prometheus.Labels{
			"path":ctx.Path(),
			"method": ctx.Request().Method,
		}).Add(1)

		return fn(ctx)
	}
}

func Pull() {
	svc := echo.New()
	svc.Use(pullMiddleware)
	svc.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	svc.GET("/test1", TestHandler)
	svc.GET("/test2", TestHandler)
	if err := svc.Start(":9090"); err != nil {
		panic(err)
	}
}
