package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/model"
)

type announcementRepoSpy struct {
	lastOffset int
	lastLimit  int
}

func (a *announcementRepoSpy) Create(ctx context.Context, in *model.Announcement) error {
	_, _ = ctx, in
	return nil
}

func (a *announcementRepoSpy) Update(ctx context.Context, in *model.Announcement) error {
	_, _ = ctx, in
	return nil
}

func (a *announcementRepoSpy) GetByID(ctx context.Context, id int64) (*model.Announcement, error) {
	_, _ = ctx, id
	return nil, nil
}

func (a *announcementRepoSpy) List(ctx context.Context, offset, limit int) ([]model.Announcement, int64, error) {
	_ = ctx
	a.lastOffset = offset
	a.lastLimit = limit
	return nil, 0, nil
}

func (a *announcementRepoSpy) Delete(ctx context.Context, id int64) error {
	_, _ = ctx, id
	return nil
}

func TestAnnouncementService_List_paging(t *testing.T) {
	spy := &announcementRepoSpy{}
	svc := NewAnnouncementService(spy)
	_, _, err := svc.List(context.Background(), 2, 15)
	require.NoError(t, err)
	require.Equal(t, 15, spy.lastOffset)
	require.Equal(t, 15, spy.lastLimit)
}

func TestAnnouncementService_List_normalizesDefaults(t *testing.T) {
	spy := &announcementRepoSpy{}
	svc := NewAnnouncementService(spy)
	_, _, err := svc.List(context.Background(), 0, 200)
	require.NoError(t, err)
	require.Equal(t, 0, spy.lastOffset)
	require.Equal(t, 20, spy.lastLimit)
}
