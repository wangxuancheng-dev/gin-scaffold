package job

import (
	"context"
	"testing"
)

func TestSetAuditExportStatus_invalidArgs(t *testing.T) {
	ctx := context.Background()
	if err := SetAuditExportStatus(ctx, nil); err == nil {
		t.Fatal("expected error for nil status")
	}
	if err := SetAuditExportStatus(ctx, &AuditExportStatus{}); err == nil {
		t.Fatal("expected error for empty task id")
	}
}
