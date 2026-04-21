package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/config"
	"gin-scaffold/internal/dao"
)

func TestOutboxDispatcher_Start_nilReceiver(t *testing.T) {
	var d *OutboxDispatcher
	stop := d.Start()
	require.NotPanics(t, func() { stop() })
}

func TestOutboxDispatcher_Start_disabled(t *testing.T) {
	d := NewOutboxDispatcher(&dao.OutboxDAO{}, config.OutboxConfig{Enabled: false})
	stop := d.Start()
	require.NotPanics(t, func() { stop() })
}

func TestOutboxDispatcher_Start_nilDAO(t *testing.T) {
	d := NewOutboxDispatcher(nil, config.OutboxConfig{Enabled: true})
	stop := d.Start()
	require.NotPanics(t, func() { stop() })
}
