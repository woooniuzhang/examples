package prometheus

import (
	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"log"
	"time"
)

//服务主动推的方式，一般使用于非long running的或者对监控及时性要求比较高的场景下
//单节点的pushgateway 可能会成为性能瓶颈

func TestHandler(ctx echo.Context) error {
	return ctx.String(200, "ok")
}

var counterVec *prometheus.CounterVec
var histogramVec *prometheus.HistogramVec

func init() {
	//注意第二个参数,里面指定的label在Add的时候必须要体现到，否则会panic
	counterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_total",
		ConstLabels: prometheus.Labels{
			"svcName": "echo-test",
		},
		Help:"request total",
	}, []string{"path","method"})

	histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_request_duration",
		Help:        "http_request_duration",
		ConstLabels: prometheus.Labels{
			"svcName": "echo-test",
		},
		Buckets:[]float64{10, 100, 500, 1000},
	}, []string{"path", "method"})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("push panic: %v", r)
			}
		}()
		timer := time.NewTimer(time.Second * 5)

		// docker pull prom/pushgateway
		// docker run -it -p 9091:9091 prom/pushgateway
		httpRequestPusher := push.New("http://127.0.0.1:9091", "http_request").Collector(counterVec)
		httpRequestDurationPusher := push.New("http://127.0.0.1:9091", "http_request_duration").Collector(histogramVec)
		for {
			select {
			case <-timer.C:
				if err := httpRequestPusher.Push(); err != nil {
					log.Printf("push metric error: %v", err)
				}

				if err := httpRequestDurationPusher.Push(); err != nil {
					log.Panicf("push duration metric error: %v", err)
				}
				timer.Reset(time.Second * 5)
			}
		}
	}()
}

func PrometheusMiddleware(fn echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		now := time.Now()

		defer func() {
			if err := recover(); err != nil {
				log.Printf("middleware panic: %v", err)
			}
		}()

		counterVec.With(prometheus.Labels{
			"path":ctx.Path(),
			"method": ctx.Request().Method,
		}).Add(1)

		histogramVec.With(prometheus.Labels{
			"path": ctx.Path(),
			"method": ctx.Request().Method,
		}).Observe(float64(time.Since(now).Milliseconds()))


		return fn(ctx)
	}
}

func Push() {
	server := echo.New()
	server.GET("/test1", TestHandler)
	server.GET("/test2", TestHandler)
	server.Use(PrometheusMiddleware)
	if err := server.Start(":9090"); err != nil {
		panic(err)
	}
}
