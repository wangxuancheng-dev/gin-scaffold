package clientresp

import (
	"testing"
	"time"

	"gin-scaffold/internal/model"
)

func TestFromUser_nil(t *testing.T) {
	vo := FromUser(nil)
	if vo.ID != 0 || vo.Username != "" {
		t.Fatalf("%+v", vo)
	}
}

func TestFromUser_formatsTimes(t *testing.T) {
	ts := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	u := &model.User{
		ID:        1,
		Username:  "a",
		Nickname:  "n",
		CreatedAt: ts,
		UpdatedAt: ts,
	}
	vo := FromUser(u)
	if vo.CreatedAt == "" || vo.UpdatedAt == "" {
		t.Fatalf("%+v", vo)
	}
}
