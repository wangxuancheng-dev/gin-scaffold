package middleware

import (
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// Metrics 注册 Prometheus 中间件（使用默认注册表）。
func Metrics(engine *gin.Engine, subsystem, path string) {
	if subsystem == "" {
		subsystem = "http"
	}
	if path == "" {
		path = "/metrics"
	}
	p := ginprometheus.NewPrometheus(subsystem)
	p.MetricsPath = path
	p.Use(engine)
}
