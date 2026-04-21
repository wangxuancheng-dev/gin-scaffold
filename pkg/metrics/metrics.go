// Package metrics 暴露 Prometheus 注册表与 Asynq 相关计数扩展点。
package metrics

import (
	"time"

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

// AsynqTaskDurationSeconds 异步任务处理耗时分布。
var AsynqTaskDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "asynq_task_duration_seconds",
		Help:    "Duration of asynq task processing in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"type", "status"},
)

// OutboxEventsTotal Outbox 事件生命周期计数（发布/重试/死信）。
var OutboxEventsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "outbox_events_total",
		Help: "Number of outbox events by topic and status",
	},
	[]string{"topic", "status"},
)

// OutboxRetryDelaySeconds Outbox 下一次重试延迟分布。
var OutboxRetryDelaySeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "outbox_retry_delay_seconds",
		Help:    "Delay before next outbox retry in seconds",
		Buckets: []float64{1, 3, 5, 10, 30, 60, 120, 300},
	},
	[]string{"topic"},
)

// SchedulerTaskExecutionsTotal 定时任务执行结果计数。
var SchedulerTaskExecutionsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "scheduler_task_executions_total",
		Help: "Number of scheduler task executions by status",
	},
	[]string{"status"},
)

// SchedulerTaskDurationSeconds 定时任务执行耗时分布。
var SchedulerTaskDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "scheduler_task_duration_seconds",
		Help:    "Duration of scheduler task executions in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"status"},
)

// ObserveAsynqTask records both count and duration in one place.
func ObserveAsynqTask(taskType, status string, start time.Time) {
	AsynqTasksProcessed.WithLabelValues(taskType, status).Inc()
	if !start.IsZero() {
		AsynqTaskDurationSeconds.WithLabelValues(taskType, status).Observe(time.Since(start).Seconds())
	}
}
