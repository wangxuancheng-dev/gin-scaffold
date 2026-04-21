package adminhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	adminreq "gin-scaffold/internal/api/request/admin"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/service/port"
)

type stubAnnouncementService struct {
	listRows []model.Announcement
	listTot  int64
	listErr  error

	getRow *model.Announcement
	getErr error

	createErr error
	updateErr error
	deleteErr error
}

func (s *stubAnnouncementService) Create(ctx context.Context, in *model.Announcement) error {
	return s.createErr
}

func (s *stubAnnouncementService) Update(ctx context.Context, in *model.Announcement) error {
	return s.updateErr
}

func (s *stubAnnouncementService) GetByID(ctx context.Context, id int64) (*model.Announcement, error) {
	return s.getRow, s.getErr
}

func (s *stubAnnouncementService) List(ctx context.Context, page, pageSize int) ([]model.Announcement, int64, error) {
	return s.listRows, s.listTot, s.listErr
}

func (s *stubAnnouncementService) Delete(ctx context.Context, id int64) error {
	return s.deleteErr
}

var _ port.AnnouncementService = (*stubAnnouncementService)(nil)

func TestAnnouncementHandler_List_ok(t *testing.T) {
	svc := &stubAnnouncementService{
		listRows: []model.Announcement{{ID: 1, Title: "T", Content: "C", Status: "draft"}},
		listTot:  1,
	}
	h := NewAnnouncementHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/announcements?page=1&page_size=10", nil)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAnnouncementHandler_Get_notFound(t *testing.T) {
	svc := &stubAnnouncementService{getErr: gorm.ErrRecordNotFound}
	h := NewAnnouncementHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/announcements/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	h.Get(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAnnouncementHandler_Create_ok(t *testing.T) {
	svc := &stubAnnouncementService{}
	h := NewAnnouncementHandler(svc)
	body, _ := json.Marshal(adminreq.AnnouncementCreateRequest{Title: "T", Content: "C", Status: "draft"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/admin/announcements", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAnnouncementHandler_Update_ok(t *testing.T) {
	svc := &stubAnnouncementService{}
	h := NewAnnouncementHandler(svc)
	title := "T2"
	body, _ := json.Marshal(adminreq.AnnouncementUpdateRequest{Title: &title})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "5"}}
	c.Request = httptest.NewRequest(http.MethodPut, "http://localhost/admin/announcements/5", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Update(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAnnouncementHandler_Delete_ok(t *testing.T) {
	svc := &stubAnnouncementService{}
	h := NewAnnouncementHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "http://localhost/admin/announcements/3", nil)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	h.Delete(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestAnnouncementHandler_List_serviceError(t *testing.T) {
	svc := &stubAnnouncementService{listErr: errors.New("db")}
	h := NewAnnouncementHandler(svc)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/admin/announcements", nil)
	h.List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", w.Code)
	}
}
