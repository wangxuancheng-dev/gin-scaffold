package job

import (
	"context"
	"testing"
)

func TestSetUserExportStatus_invalidArgs(t *testing.T) {
	ctx := context.Background()
	if err := SetUserExportStatus(ctx, nil); err == nil {
		t.Fatal("expected error for nil status")
	}
	if err := SetUserExportStatus(ctx, &UserExportStatus{}); err == nil {
		t.Fatal("expected error for empty task id")
	}
}
