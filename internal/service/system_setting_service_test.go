package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/model"
)

// systemSettingRepoSpy records List pagination args.
type systemSettingRepoSpy struct {
	lastOffset int
	lastLimit  int
}

func (s *systemSettingRepoSpy) Create(ctx context.Context, row *model.SystemSetting) error {
	_, _ = ctx, row
	return nil
}

func (s *systemSettingRepoSpy) GetByID(ctx context.Context, id int64) (*model.SystemSetting, error) {
	_, _ = ctx, id
	return nil, nil
}

func (s *systemSettingRepoSpy) GetByIDAny(ctx context.Context, id int64) (*model.SystemSetting, error) {
	_, _ = ctx, id
	return nil, nil
}

func (s *systemSettingRepoSpy) GetByKey(ctx context.Context, key string) (*model.SystemSetting, error) {
	_, _ = ctx, key
	return nil, nil
}

func (s *systemSettingRepoSpy) List(ctx context.Context, q model.SystemSettingQuery, offset, limit int) ([]model.SystemSetting, int64, error) {
	_, _ = ctx, q
	s.lastOffset = offset
	s.lastLimit = limit
	return nil, 0, nil
}

func (s *systemSettingRepoSpy) Update(ctx context.Context, id int64, updates map[string]any) (*model.SystemSetting, error) {
	_, _, _ = ctx, id, updates
	return nil, nil
}

func (s *systemSettingRepoSpy) Restore(ctx context.Context, id int64, updates map[string]any) (*model.SystemSetting, error) {
	_, _, _ = ctx, id, updates
	return nil, nil
}

func (s *systemSettingRepoSpy) SoftDelete(ctx context.Context, id int64) error {
	_, _ = ctx, id
	return nil
}

func (s *systemSettingRepoSpy) CreateHistory(ctx context.Context, row *model.SystemSettingHistory) error {
	_, _ = ctx, row
	return nil
}

func (s *systemSettingRepoSpy) GetHistoryByID(ctx context.Context, id int64) (*model.SystemSettingHistory, error) {
	_, _ = ctx, id
	return nil, nil
}

func (s *systemSettingRepoSpy) ListHistory(ctx context.Context, settingID int64, offset, limit int) ([]model.SystemSettingHistory, int64, error) {
	_, _, _, _ = ctx, settingID, offset, limit
	return nil, 0, nil
}

func TestSystemSettingService_List_normalizesPaging(t *testing.T) {
	spy := &systemSettingRepoSpy{}
	svc := NewSystemSettingService(spy)
	_, _, err := svc.List(context.Background(), model.SystemSettingQuery{}, 3, 10)
	require.NoError(t, err)
	require.Equal(t, 20, spy.lastOffset)
	require.Equal(t, 10, spy.lastLimit)
}

func TestSystemSettingService_List_coercesPageSizeCap(t *testing.T) {
	spy := &systemSettingRepoSpy{}
	svc := NewSystemSettingService(spy)
	_, _, err := svc.List(context.Background(), model.SystemSettingQuery{}, 1, 500)
	require.NoError(t, err)
	require.Equal(t, 0, spy.lastOffset)
	require.Equal(t, 20, spy.lastLimit)
}

func TestSystemSettingService_List_pageBelowOne(t *testing.T) {
	spy := &systemSettingRepoSpy{}
	svc := NewSystemSettingService(spy)
	_, _, err := svc.List(context.Background(), model.SystemSettingQuery{}, 0, 5)
	require.NoError(t, err)
	require.Equal(t, 0, spy.lastOffset)
	require.Equal(t, 5, spy.lastLimit)
}
