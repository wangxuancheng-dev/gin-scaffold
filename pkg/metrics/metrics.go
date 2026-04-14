// Package metrics 暴露 Prometheus 注册表与 Asynq 相关计数扩展点。
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry 默认使用 prometheus 默认注册表（与 gin 中间件兼容）。
var Registry = prometheus.DefaultRegisterer

// AsynqTasksProcessed 任务处理计数（业务 handler 内可 Inc）。
var AsynqTasksProcessed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "asynq_tasks_processed_total",
		Help: "Number of asynq tasks processed",
	},
	[]string{"type", "status"},
)
