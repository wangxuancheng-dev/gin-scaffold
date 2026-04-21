package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFields_BasicAndOptional(t *testing.T) {
	t.Parallel()

	fields, err := parseFields([]string{
		"title:string:required,max=64",
		"note:string?:max=255",
		"status:string:oneof=draft published,default=draft",
		"amount:int:min=0,default=1",
	})
	if err != nil {
		t.Fatalf("parseFields error: %v", err)
	}
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(fields))
	}

	if fields[1].Optional != true {
		t.Fatalf("expected note optional=true")
	}
	if fields[2].Default != "draft" {
		t.Fatalf("expected status default=draft, got %q", fields[2].Default)
	}
	if fields[3].Validate != "min=0" {
		t.Fatalf("expected amount validate=min=0, got %q", fields[3].Validate)
	}
}

func TestParseFields_DuplicateName(t *testing.T) {
	t.Parallel()
	_, err := parseFields([]string{"title:string", "title:int"})
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestParseValidateAndDefault(t *testing.T) {
	t.Parallel()
	v, d := parseValidateAndDefault("oneof=open closed,default=open")
	if v != "oneof=open closed" || d != "open" {
		t.Fatalf("unexpected parse result validate=%q default=%q", v, d)
	}
}

func TestRequestTemplate_OptionalAndValidate(t *testing.T) {
	t.Parallel()
	src := requestTemplate("Ticket", []genField{
		{Name: "Title", JSONName: "title", GoType: "string", Validate: "required,max=64"},
		{Name: "Note", JSONName: "note", GoType: "string", Optional: true, Validate: "max=255"},
	})
	if !strings.Contains(src, "binding:\"required,max=64\"") {
		t.Fatalf("missing create required binding: %s", src)
	}
	if !strings.Contains(src, "binding:\"max=255\"") {
		t.Fatalf("missing create optional binding: %s", src)
	}
	if !strings.Contains(src, "binding:\"omitempty,max=255\"") {
		t.Fatalf("missing update optional binding: %s", src)
	}
}

func TestSchemaUpTemplate_Defaults(t *testing.T) {
	t.Parallel()
	src := schemaUpTemplate("tickets", []genField{
		{Name: "Status", JSONName: "status", GoType: "string", SQLType: "VARCHAR(32)", Default: "open"},
		{Name: "Paid", JSONName: "paid", GoType: "bool", SQLType: "TINYINT(1)", Default: "true"},
		{Name: "Priority", JSONName: "priority", GoType: "int", SQLType: "INT", Default: "3"},
	})
	wants := []string{
		"`tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default'",
		"`status` VARCHAR(32) NOT NULL DEFAULT 'open'",
		"`paid` TINYINT(1) NOT NULL DEFAULT 1",
		"`priority` INT NOT NULL DEFAULT 3",
	}
	for _, want := range wants {
		if !strings.Contains(src, want) {
			t.Fatalf("schema should contain %q, got: %s", want, src)
		}
	}
}

func TestModelTemplate_ContainsTenantField(t *testing.T) {
	t.Parallel()
	src := modelTemplate("Ticket", "tickets", []genField{
		{Name: "Title", JSONName: "title", GoType: "string"},
	})
	if !strings.Contains(src, "TenantID  string") {
		t.Fatalf("model template should contain TenantID field: %s", src)
	}
	if !strings.Contains(src, "json:\"tenant_id\"") {
		t.Fatalf("model template should expose tenant_id json tag: %s", src)
	}
}

func TestDAOTemplate_ContainsTenantScope(t *testing.T) {
	t.Parallel()
	src := daoTemplate("Ticket", "TicketDAO")
	wants := []string{
		"internal/pkg/tenant",
		"in.TenantID = tenant.FromContext(ctx)",
		"tenant.ApplyScope(ctx",
	}
	for _, want := range wants {
		if !strings.Contains(src, want) {
			t.Fatalf("dao template should contain %q, got: %s", want, src)
		}
	}
}

func TestNormalizeCrudOptions_SimpleForcesNoWire(t *testing.T) {
	t.Parallel()
	opt := &crudOptions{
		module:   "demo",
		template: "simple",
		outDir:   ".",
	}
	forced, err := normalizeCrudOptions(opt)
	if err != nil {
		t.Fatalf("normalize error: %v", err)
	}
	if !forced || !opt.noWire {
		t.Fatalf("expected simple template to force no-wire, forced=%v noWire=%v", forced, opt.noWire)
	}
	if opt.table != "demos" {
		t.Fatalf("expected default table demos, got %q", opt.table)
	}
}

func TestNormalizeCrudOptions_OutDirGuard(t *testing.T) {
	t.Parallel()
	opt := &crudOptions{
		module:   "demo",
		template: "full",
		outDir:   filepath.Join("tmp", "preview"),
	}
	_, err := normalizeCrudOptions(opt)
	if err == nil {
		t.Fatal("expected out-dir guard error")
	}
}

func TestWireAdminRouter_CreatesRegisterWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir tmp: %v", err)
	}

	if err := wireAdminRouter("Demo"); err != nil {
		t.Fatalf("wireAdminRouter error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join("routes", "adminroutes", "register.go"))
	if err != nil {
		t.Fatalf("read register.go: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "generatedDemo *adminhandler.DemoHandler") {
		t.Fatalf("missing generated demo param in register.go: %s", text)
	}
	if !strings.Contains(text, "registerAdminDemoRoutes(admin, generatedDemo)") {
		t.Fatalf("missing generated demo route registration in register.go: %s", text)
	}
}

func TestAdminHandlerTemplate_UsesUnifiedErrorHelpers(t *testing.T) {
	t.Parallel()
	src := adminHandlerTemplate("Ticket", "TicketService", []genField{
		{Name: "Title", JSONName: "title", GoType: "string"},
	})
	wants := []string{
		"handler.FailInvalidParam(c, err)",
		"handler.FailInternal(c, err)",
		"errors.Is(err, gorm.ErrRecordNotFound)",
		"handler.FailNotFound(c, \"\")",
	}
	for _, want := range wants {
		if !strings.Contains(src, want) {
			t.Fatalf("handler template should contain %q, got: %s", want, src)
		}
	}
}

func TestAdminRouteTemplate_UsesSnakeCaseResource(t *testing.T) {
	t.Parallel()
	src := adminRouteTemplate("OrderItem", "order_item")
	if !strings.Contains(src, "/order_items") {
		t.Fatalf("route template should use snake_case resource name, got: %s", src)
	}
	if !strings.Contains(src, "order_item:read") || !strings.Contains(src, "order_item:write") {
		t.Fatalf("route template should use snake_case permission prefix, got: %s", src)
	}
}
