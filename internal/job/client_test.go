package job

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestClient_EnqueueTaskInQueueAfter_scheduled(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(func() { mr.Close() })

	cli := NewClient(mr.Addr(), "", 0, "default", 0, 60, 0)
	ctx := context.Background()
	payload := WelcomePayload{UserID: 1, Username: "u1"}

	require.NoError(t, cli.EnqueueTaskInQueueAfter(ctx, "default", TypeWelcomeEmail, payload, 0, 30*time.Minute))

	ins := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr(), DB: 0})
	info, err := ins.GetQueueInfo("default")
	require.NoError(t, err)
	require.GreaterOrEqual(t, info.Scheduled, 1, "delayed task should appear in scheduled bucket")
}

func TestClient_EnqueueTaskInQueueAfter_zeroDelay_immediate(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(func() { mr.Close() })

	cli := NewClient(mr.Addr(), "", 0, "default", 0, 60, 0)
	ctx := context.Background()
	payload := WelcomePayload{UserID: 2, Username: "u2"}

	require.NoError(t, cli.EnqueueTaskInQueueAfter(ctx, "default", TypeWelcomeEmail, payload, 0, 0))

	ins := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr(), DB: 0})
	info, err := ins.GetQueueInfo("default")
	require.NoError(t, err)
	require.GreaterOrEqual(t, info.Pending, 1)
}

func TestClient_EnqueueTaskInQueueAt_future(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(func() { mr.Close() })

	cli := NewClient(mr.Addr(), "", 0, "low", 0, 60, 0)
	ctx := context.Background()
	payload := WelcomePayload{UserID: 3, Username: "u3"}
	at := time.Now().Add(15 * time.Minute)

	require.NoError(t, cli.EnqueueTaskInQueueAt(ctx, "low", TypeWelcomeEmail, payload, 0, at))

	ins := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr(), DB: 0})
	info, err := ins.GetQueueInfo("low")
	require.NoError(t, err)
	require.GreaterOrEqual(t, info.Scheduled, 1)
}
