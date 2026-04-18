package model

import "testing"

func TestUser_TableName(t *testing.T) {
	if (&User{}).TableName() != "users" {
		t.Fatal()
	}
}
