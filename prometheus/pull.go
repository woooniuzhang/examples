package prometheus

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"time"
)

//经典的pull模式, 适合long running 或者对监控实时性要求不是特别高的场景。

var pullCounterVec *prometheus.CounterVec
var pullHistogramVec *prometheus.HistogramVec

func init() {
	pullCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_total",
		Help: "http_request_total",
	}, []string{"method","path"})

	pullHistogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_request_duration",
		Help:        "http_request_duration",
		Buckets: []float64{10, 100, 500, 1000},
	}, []string{"method", "path"})

	//将此metric信息注册进Gather中，这样在提供metrics接口时就可以
	//按照prometheus的方式将metric展现出来了
	prometheus.MustRegister(pullCounterVec)
	prometheus.MustRegister(pullHistogramVec)
}

func Middleware(fn echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		now := time.Now()

		defer func() {
			if err := recover(); err != nil {
				log.Printf("middleware panic: %v", err)
			}
		}()

		pullCounterVec.With(prometheus.Labels{
			"path":ctx.Path(),
			"method": ctx.Request().Method,
		}).Add(1)

		pullHistogramVec.With(prometheus.Labels{
			"path": ctx.Path(),
			"method": ctx.Request().Method,
		}).Observe(float64(time.Since(now).Milliseconds()))

		return fn(ctx)
	}
}

func Pull() {
	svc := echo.New()
	svc.Use(Middleware)
	svc.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	svc.GET("/test1", TestHandler)
	svc.GET("/test2", TestHandler)
	if err := svc.Start(":9090"); err != nil {
		panic(err)
	}
}
