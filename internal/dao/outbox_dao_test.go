package dao

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/pkg/tenant"
)

func TestOutboxDAO_Enqueue(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	ctx := tenant.WithContext(context.Background(), "t1")
	mock.ExpectExec("INSERT INTO `outbox_events`").WillReturnResult(sqlmock.NewResult(1, 1))
	ev, err := d.Enqueue(ctx, "orders.created", `{"k":1}`, 7)
	require.NoError(t, err)
	require.NotNil(t, ev)
	require.Equal(t, "t1", ev.TenantID)
	require.Equal(t, "orders.created", ev.Topic)
	require.Equal(t, 7, ev.MaxAttempts)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxDAO_Enqueue_defaultTenant(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	mock.ExpectExec("INSERT INTO `outbox_events`").WillReturnResult(sqlmock.NewResult(2, 1))
	ev, err := d.Enqueue(context.Background(), "t", "p", 10)
	require.NoError(t, err)
	require.Equal(t, "default", ev.TenantID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxDAO_EnqueueTx_nilDelegatesToEnqueue(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	mock.ExpectExec("INSERT INTO `outbox_events`").WillReturnResult(sqlmock.NewResult(3, 1))
	ev, err := d.EnqueueTx(context.Background(), nil, "t", "p", 5)
	require.NoError(t, err)
	require.Equal(t, "default", ev.TenantID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxDAO_FetchDue_limitDefault(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "tenant_id", "topic", "payload", "status", "attempts", "max_attempts",
		"next_run_at", "last_error", "published_at", "created_at", "updated_at",
	}).AddRow(int64(1), "default", "t", "{}", "pending", 0, 10, now, "", nil, now, now)
	mock.ExpectQuery("SELECT .* FROM `outbox_events`").WillReturnRows(rows)
	got, err := d.FetchDue(context.Background(), 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "pending", got[0].Status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxDAO_MarkPublished(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	mock.ExpectExec("UPDATE `outbox_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, d.MarkPublished(context.Background(), 9))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxDAO_MarkRetry(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	mock.ExpectExec("UPDATE `outbox_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	next := time.Now().Add(time.Minute)
	require.NoError(t, d.MarkRetry(context.Background(), 1, 2, next, "retry me"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxDAO_MarkDead(t *testing.T) {
	db, mock := newTestDB(t)
	d := NewOutboxDAO(db)
	mock.ExpectExec("UPDATE `outbox_events`").WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, d.MarkDead(context.Background(), 1, 10, "gone"))
	require.NoError(t, mock.ExpectationsWereMet())
}
