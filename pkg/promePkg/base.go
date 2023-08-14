package promePkg

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ApiRequestCounter        prometheus.Counter
	ApiRequestSuccessCounter prometheus.Counter
	ApiRequestFailedCounter  prometheus.Counter

	AccountConnCounter    prometheus.Counter
	AccountDisConnCounter prometheus.Counter

	PrometheusEnable bool = true
)

func NewAccountDisConnCounter() {
	if AccountDisConnCounter != nil {
		return
	}
	AccountDisConnCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "account_dis_conn",
		Help: "账户断开链接数",
	})
}

func NewAccountConnCounter() {
	if AccountConnCounter != nil {
		return
	}
	AccountConnCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "account_conn",
		Help: "账户链接数",
	})
}

func NewApiRequestCounter() {
	if ApiRequestCounter != nil {
		return
	}
	ApiRequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_request",
		Help: "The number of api request",
	})
}

func NewApiRequestSuccessCounter() {
	if ApiRequestSuccessCounter != nil {
		return
	}
	ApiRequestSuccessCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_request_success",
		Help: "The number of api request success",
	})
}

func NewApiRequestFailedCounter() {
	if ApiRequestFailedCounter != nil {
		return
	}
	ApiRequestFailedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_request_failed",
		Help: "The number of api request failed",
	})
}

func InitConter() {
	NewApiRequestCounter()
	NewApiRequestFailedCounter()
	NewApiRequestSuccessCounter()

	NewAccountConnCounter()
	NewAccountDisConnCounter()

}

func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func PromeInc(counter prometheus.Counter) {
	if PrometheusEnable {
		if counter != nil {
			counter.Inc()
		}
	}
}

func PromeAdd(counter prometheus.Counter, add int) {
	if PrometheusEnable {
		if counter != nil {
			counter.Add(float64(add))
		}
	}
}

func PromeGaugeInc(gauges prometheus.Gauge) {
	if PrometheusEnable {
		if gauges != nil {
			gauges.Inc()
		}
	}
}

func PromeGaugeDec(gauges prometheus.Gauge) {
	if PrometheusEnable {
		if gauges != nil {
			gauges.Dec()
		}
	}
}
