package bootstrap

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/config"
)

func TestServerDeps_Cleanup_nilReceiver(t *testing.T) {
	var d *ServerDeps
	d.Cleanup(context.Background())
}

func TestServerDeps_Cleanup_nilCleanupFunc(t *testing.T) {
	d := &ServerDeps{}
	d.Cleanup(context.Background())
}

func TestServerDeps_Cleanup_invokesInOrder(t *testing.T) {
	var order []int
	d := &ServerDeps{
		cleanup: func(context.Context) {
			order = append(order, 1)
		},
	}
	d.Cleanup(context.Background())
	require.Equal(t, []int{1}, order)
}

func TestWorkerDeps_Cleanup_nilReceiver(t *testing.T) {
	var d *WorkerDeps
	d.Cleanup(context.Background())
}

func TestWorkerDeps_Cleanup_nilCleanupFunc(t *testing.T) {
	d := &WorkerDeps{}
	d.Cleanup(context.Background())
}

func TestRunCleanups_reverseOrder(t *testing.T) {
	var order []int
	runCleanups(context.Background(), []func(context.Context){
		func(context.Context) { order = append(order, 1) },
		func(context.Context) { order = append(order, 2) },
		func(context.Context) { order = append(order, 3) },
	})
	require.Equal(t, []int{3, 2, 1}, order)
}

func TestResolveAsynqQueues_explicitMap(t *testing.T) {
	cfg := config.AsynqConfig{Queues: map[string]int{"hi": 2, "lo": 1}}
	got := resolveAsynqQueues(cfg)
	require.Equal(t, map[string]int{"hi": 2, "lo": 1}, got)
}

func TestResolveAsynqQueues_fallbackQueueName(t *testing.T) {
	cfg := config.AsynqConfig{Queue: "jobs"}
	require.Equal(t, map[string]int{"jobs": 1}, resolveAsynqQueues(cfg))
}

func TestResolveAsynqQueues_defaultQueue(t *testing.T) {
	cfg := config.AsynqConfig{}
	require.Equal(t, map[string]int{"default": 1}, resolveAsynqQueues(cfg))
}
