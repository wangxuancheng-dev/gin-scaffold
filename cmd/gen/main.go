package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"
)

type crudOptions struct {
	module      string
	table       string
	template    string
	force       bool
	noWire      bool
	dryRun      bool
	fields      []string
	previewFile string
	previewFull bool
	outDir      string
}

type genField struct {
	Name     string
	JSONName string
	GoType   string
	SQLType  string
	Validate string
	Optional bool
	Default  string
}

func main() {
	var opt crudOptions
	rootCmd := &cobra.Command{
		Use:   "gen",
		Short: "code generator",
		Run: func(cmd *cobra.Command, args []string) {
			usage()
		},
	}
	crudCmd := &cobra.Command{
		Use:   "crud",
		Short: "generate CRUD scaffold",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runCRUD(opt); err != nil {
				return fmt.Errorf("crud generator error: %w", err)
			}
			return nil
		},
	}
	crudCmd.Flags().StringVar(&opt.module, "module", "", "module name, e.g. order")
	crudCmd.Flags().StringVar(&opt.table, "table", "", "table name, default: <module>s")
	crudCmd.Flags().StringVar(&opt.template, "template", "full", "generation template: full|simple")
	crudCmd.Flags().BoolVar(&opt.force, "force", false, "overwrite existing files")
	crudCmd.Flags().BoolVar(&opt.noWire, "no-wire", false, "only generate files, do not auto wire routes/bootstrap")
	crudCmd.Flags().BoolVar(&opt.dryRun, "dry-run", false, "preview files to be generated without writing")
	crudCmd.Flags().StringArrayVar(&opt.fields, "field", nil, "field definition, format: name:type[:validate][,default=value]. type can end with ? for optional create field")
	crudCmd.Flags().StringVar(&opt.previewFile, "preview-file", "", "write generation preview into a markdown file")
	crudCmd.Flags().BoolVar(&opt.previewFull, "preview-full", false, "when used with --preview-file, include full file contents")
	crudCmd.Flags().StringVar(&opt.outDir, "out-dir", ".", "output base directory for generated files")
	_ = crudCmd.MarkFlagRequired("module")
	rootCmd.AddCommand(crudCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/gen crud --module <name> [--table <table_name>] [--template full|simple] [--field name:type[:validate]] [--force] [--no-wire] [--dry-run] [--preview-file <path>] [--preview-full] [--out-dir <path>]")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("  go run ./cmd/gen crud --module order --template full --field title:string:required,max=64 --field status:string:oneof=draft published,default=draft --dry-run --preview-file ./tmp/order-preview.md --preview-full")
	fmt.Println("  go run ./cmd/gen crud --module order --no-wire --out-dir ./tmp/scaffold-preview")
}

func runCRUD(opt crudOptions) error {
	opt.module = strings.TrimSpace(opt.module)
	if opt.module == "" {
		return errors.New("module is required")
	}
	templateForcesNoWire, err := normalizeCrudOptions(&opt)
	if err != nil {
		return err
	}
	parsedFields, err := parseFields(opt.fields)
	if err != nil {
		return err
	}

	modelName := toPascal(opt.module)
	moduleSnake := toSnake(opt.module)
	daoName := modelName + "DAO"
	serviceName := modelName + "Service"
	permPrefix := moduleSnake
	now := time.Now().Format("200601021504")
	migrationBase := fmt.Sprintf("%s_create_%s", now, opt.table)
	seedBase := fmt.Sprintf("%s_seed_%s_permission", now, moduleSnake)

	files := map[string]string{
		filepath.Join("internal", "model", moduleSnake+".go"):                   modelTemplate(modelName, opt.table, parsedFields),
		filepath.Join("internal", "dao", moduleSnake+"_dao.go"):                 daoTemplate(modelName, daoName),
		filepath.Join("internal", "service", "port", moduleSnake+"_service.go"): portTemplate(modelName, serviceName),
		filepath.Join("internal", "service", moduleSnake+"_service.go"):         serviceTemplate(modelName, serviceName),
		filepath.Join("api", "request", "admin", moduleSnake+"_request.go"):     requestTemplate(modelName, parsedFields),
		filepath.Join("api", "handler", "admin", moduleSnake+"_handler.go"):     adminHandlerTemplate(modelName, serviceName, parsedFields),
		filepath.Join("routes", "adminroutes", moduleSnake+"_router.go"):        adminRouteTemplate(modelName),
	}
	if opt.template == "full" {
		files[filepath.Join("migrations", "mysql", "schema", migrationBase+".up.sql")] = schemaUpTemplate(opt.table, parsedFields)
		files[filepath.Join("migrations", "mysql", "schema", migrationBase+".down.sql")] = schemaDownTemplate(opt.table)
		files[filepath.Join("migrations", "mysql", "seed", seedBase+".up.sql")] = seedPermUpTemplate(permPrefix)
		files[filepath.Join("migrations", "mysql", "seed", seedBase+".down.sql")] = seedPermDownTemplate(permPrefix)
	}

	for p, content := range files {
		if err := writeFile(opt.outDir, p, content, opt.force, opt.dryRun); err != nil {
			return err
		}
	}
	if err := writePreview(opt.previewFile, opt, files); err != nil {
		return err
	}

	if !opt.noWire && opt.template == "full" {
		if err := wireGeneratedCRUD(moduleSnake, modelName, daoName, serviceName); err != nil {
			return err
		}
	}

	if opt.dryRun {
		fmt.Println("Dry run only. Planned files:")
	} else {
		fmt.Println("CRUD scaffold generated:")
	}
	if templateForcesNoWire {
		fmt.Println("Note: template=simple forces --no-wire mode.")
	}
	for p := range files {
		fmt.Println(" -", p)
	}
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1) review request validation tags and generated handler field mapping")
	if opt.template == "full" {
		fmt.Println("  2) review generated migration/seed SQL for table:", opt.table)
		fmt.Println("  3) apply migration: go run ./cmd/migrate up --env dev")
	} else {
		fmt.Println("  2) simple template skips migration/seed and auto-wiring")
	}
	if opt.noWire || opt.template == "simple" {
		fmt.Println("  3) wire route/bootstrap manually")
	}

	return nil
}

func normalizeCrudOptions(opt *crudOptions) (templateForcesNoWire bool, err error) {
	if opt == nil {
		return false, errors.New("nil options")
	}
	if opt.table == "" {
		opt.table = toSnake(opt.module) + "s"
	}
	opt.template = strings.ToLower(strings.TrimSpace(opt.template))
	if opt.template == "" {
		opt.template = "full"
	}
	if opt.template != "full" && opt.template != "simple" {
		return false, fmt.Errorf("invalid --template %q, expected full|simple", opt.template)
	}
	if opt.template == "simple" && !opt.noWire {
		opt.noWire = true
		templateForcesNoWire = true
	}
	opt.outDir = strings.TrimSpace(opt.outDir)
	if opt.outDir == "" {
		opt.outDir = "."
	}
	if filepath.Clean(opt.outDir) != "." && !opt.noWire {
		return false, errors.New("--out-dir with non-root path requires --no-wire to avoid modifying current project wiring")
	}
	return templateForcesNoWire, nil
}

func writeFile(outDir, relPath, content string, force bool, dryRun bool) error {
	abs := filepath.Join(filepath.Clean(outDir), filepath.Clean(relPath))
	if dryRun {
		return nil
	}
	if _, err := os.Stat(abs); err == nil && !force {
		return fmt.Errorf("file exists: %s (use --force to overwrite)", relPath)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}

func writePreview(previewFile string, opt crudOptions, files map[string]string) error {
	previewFile = strings.TrimSpace(previewFile)
	if previewFile == "" {
		return nil
	}
	keys := make([]string, 0, len(files))
	for p := range files {
		keys = append(keys, p)
	}
	sort.Strings(keys)
	var out strings.Builder
	out.WriteString("# CRUD Generation Preview\n\n")
	out.WriteString(fmt.Sprintf("- module: `%s`\n", opt.module))
	out.WriteString(fmt.Sprintf("- table: `%s`\n", opt.table))
	out.WriteString(fmt.Sprintf("- template: `%s`\n", opt.template))
	out.WriteString(fmt.Sprintf("- dry_run: `%t`\n", opt.dryRun))
	out.WriteString(fmt.Sprintf("- no_wire: `%t`\n", opt.noWire))
	out.WriteString(fmt.Sprintf("- preview_full: `%t`\n\n", opt.previewFull))
	out.WriteString("## Planned Files\n\n")
	for _, p := range keys {
		out.WriteString("- `" + p + "`\n")
	}
	out.WriteString("\n## File Snippets\n")
	for _, p := range keys {
		content := files[p]
		if !opt.previewFull && len(content) > 600 {
			content = content[:600] + "\n// ... truncated ..."
		}
		out.WriteString("\n### `" + p + "`\n\n```text\n" + content + "\n```\n")
	}
	if err := os.MkdirAll(filepath.Dir(previewFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(previewFile, []byte(out.String()), 0o644)
}

func wireGeneratedCRUD(moduleSnake, modelName, daoName, serviceName string) error {
	if err := wireRouterOptions(modelName); err != nil {
		return err
	}
	if err := wireAPIRouter(modelName); err != nil {
		return err
	}
	if err := wireAdminRouter(modelName); err != nil {
		return err
	}
	if err := wireBootstrap(moduleSnake, modelName, daoName, serviceName); err != nil {
		return err
	}
	return nil
}

func wireRouterOptions(modelName string) error {
	path := filepath.Join("routes", "router.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)

	fieldLine := fmt.Sprintf("\tAdmin%s  *adminhandler.%sHandler\n", modelName, modelName)
	if !strings.Contains(text, fieldLine) {
		text = strings.Replace(
			text,
			"\tAdminQueue *adminhandler.TaskQueueHandler\n",
			"\tAdminQueue *adminhandler.TaskQueueHandler\n"+fieldLine,
			1,
		)
	}

	arg := fmt.Sprintf("opts.Admin%s", modelName)
	if !strings.Contains(text, arg) {
		if strings.Contains(text, "opts.AdminSys, opts.AdminQueue, opts.AdminAnnouncement, opts.WS, opts.SSE") {
			text = strings.Replace(
				text,
				"opts.AdminSys, opts.AdminQueue, opts.AdminAnnouncement, opts.WS, opts.SSE",
				fmt.Sprintf("opts.AdminSys, opts.AdminQueue, %s, opts.AdminAnnouncement, opts.WS, opts.SSE", arg),
				1,
			)
		} else {
			text = strings.Replace(
				text,
				"opts.AdminSys, opts.AdminQueue, opts.WS, opts.SSE",
				fmt.Sprintf("opts.AdminSys, opts.AdminQueue, %s, opts.WS, opts.SSE", arg),
				1,
			)
		}
	}

	return os.WriteFile(path, []byte(text), 0o644)
}

func wireAPIRouter(modelName string) error {
	path := filepath.Join("routes", "api_router.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)

	paramLine := fmt.Sprintf("\tadmin%s *adminhandler.%sHandler,\n", modelName, modelName)
	if !strings.Contains(text, paramLine) {
		text = strings.Replace(
			text,
			"\tadminQueue *adminhandler.TaskQueueHandler,\n",
			"\tadminQueue *adminhandler.TaskQueueHandler,\n"+paramLine,
			1,
		)
	}

	arg := fmt.Sprintf("admin%s", modelName)
	if !strings.Contains(text, ", "+arg+")") {
		text = strings.Replace(
			text,
			"adminroutes.Register(r, jwtMgr, adminUser, adminMenu, adminOps, adminTask, adminSys, adminQueue, adminAnnouncement)",
			"adminroutes.Register(r, jwtMgr, adminUser, adminMenu, adminOps, adminTask, adminSys, adminQueue, "+arg+", adminAnnouncement)",
			1,
		)
	}

	return os.WriteFile(path, []byte(text), 0o644)
}

func wireAdminRouter(modelName string) error {
	path := filepath.Join("routes", "adminroutes", "register.go")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
			return mkErr
		}
		if writeErr := os.WriteFile(path, []byte(adminRegisterTemplate()), 0o644); writeErr != nil {
			return writeErr
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}
	text := string(data)

	param := fmt.Sprintf("generated%s *adminhandler.%sHandler", modelName, modelName)
	if !strings.Contains(text, param) {
		text = strings.Replace(
			text,
			"generatedAnnouncement *adminhandler.AnnouncementHandler",
			param+", generatedAnnouncement *adminhandler.AnnouncementHandler",
			1,
		)
	}

	callLine := fmt.Sprintf("\tregisterAdmin%sRoutes(admin, generated%s)\n", modelName, modelName)
	if !strings.Contains(text, callLine) {
		text = strings.Replace(text, "\tregisterAdminAnnouncementRoutes(admin, generatedAnnouncement)\n", callLine+"\tregisterAdminAnnouncementRoutes(admin, generatedAnnouncement)\n", 1)
	}

	return os.WriteFile(path, []byte(text), 0o644)
}

func wireBootstrap(moduleSnake, modelName, daoName, serviceName string) error {
	path := filepath.Join("internal", "app", "bootstrap", "bootstrap.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)

	daoLine := fmt.Sprintf("\t%sDAO := dao.New%s(gdb)\n", lowerFirst(modelName), daoName)
	if !strings.Contains(text, daoLine) {
		text = strings.Replace(text, "\tauthzDAO := dao.NewAuthzDAO(gdb)\n", daoLine+"\tauthzDAO := dao.NewAuthzDAO(gdb)\n", 1)
	}

	svcLine := fmt.Sprintf("\t%sSvc := service.New%s(%sDAO)\n", lowerFirst(modelName), serviceName, lowerFirst(modelName))
	if !strings.Contains(text, svcLine) {
		text = strings.Replace(text, "\tsysSettingSvc := service.NewSystemSettingService(sysSettingDAO)\n", "\tsysSettingSvc := service.NewSystemSettingService(sysSettingDAO)\n"+svcLine, 1)
	}

	handlerLine := fmt.Sprintf("\tadmin%sH := adminhandler.New%sHandler(%sSvc)\n", modelName, modelName, lowerFirst(modelName))
	if !strings.Contains(text, handlerLine) {
		text = strings.Replace(text, "\tadminQueueH := adminhandler.NewTaskQueueHandler(inspector)\n", "\tadminQueueH := adminhandler.NewTaskQueueHandler(inspector)\n"+handlerLine, 1)
	}

	optLine := fmt.Sprintf("\t\tAdmin%s:  admin%sH,\n", modelName, modelName)
	if !strings.Contains(text, optLine) {
		text = strings.Replace(text, "\t\tAdminQueue: adminQueueH,\n", "\t\tAdminQueue: adminQueueH,\n"+optLine, 1)
	}

	return os.WriteFile(path, []byte(text), 0o644)
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func toPascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		b.WriteRune(unicode.ToUpper(r[0]))
		b.WriteString(strings.ToLower(string(r[1:])))
	}
	return b.String()
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func parseFields(raw []string) ([]genField, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]genField, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		s := strings.TrimSpace(item)
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, ":", 3)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid --field %q, expected name:type[:validate]", s)
		}
		jsonName := toSnake(strings.TrimSpace(parts[0]))
		if jsonName == "" {
			return nil, fmt.Errorf("invalid --field %q, empty name", s)
		}
		if _, ok := seen[jsonName]; ok {
			return nil, fmt.Errorf("duplicated --field name: %s", jsonName)
		}
		typeRaw := strings.TrimSpace(parts[1])
		optional := strings.HasSuffix(typeRaw, "?")
		if optional {
			typeRaw = strings.TrimSuffix(typeRaw, "?")
		}
		goType, sqlType, ok := resolveFieldTypes(typeRaw)
		if !ok {
			return nil, fmt.Errorf("unsupported field type %q (supported: string,int,int64,bool,float64)", parts[1])
		}
		validate := ""
		defaultVal := ""
		if len(parts) == 3 {
			validate, defaultVal = parseValidateAndDefault(strings.TrimSpace(parts[2]))
		}
		out = append(out, genField{
			Name:     toPascal(jsonName),
			JSONName: jsonName,
			GoType:   goType,
			SQLType:  sqlType,
			Validate: validate,
			Optional: optional,
			Default:  defaultVal,
		})
		seen[jsonName] = struct{}{}
	}
	return out, nil
}

func resolveFieldTypes(t string) (goType string, sqlType string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "string":
		return "string", "VARCHAR(255)", true
	case "int":
		return "int", "INT", true
	case "int64":
		return "int64", "BIGINT", true
	case "bool":
		return "bool", "TINYINT(1)", true
	case "float64":
		return "float64", "DOUBLE", true
	default:
		return "", "", false
	}
}

func parseValidateAndDefault(in string) (validate string, defaultValue string) {
	s := strings.TrimSpace(in)
	if s == "" {
		return "", ""
	}
	parts := strings.Split(s, ",")
	rules := make([]string, 0, len(parts))
	for _, p := range parts {
		r := strings.TrimSpace(p)
		if r == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(r), "default=") {
			defaultValue = strings.TrimSpace(strings.TrimPrefix(r, "default="))
			continue
		}
		rules = append(rules, r)
	}
	return strings.Join(rules, ","), defaultValue
}

func modelTemplate(modelName, table string, fields []genField) string {
	var fieldLines strings.Builder
	for _, f := range fields {
		fieldLines.WriteString(fmt.Sprintf("\t%s %s `gorm:\"not null\" json:\"%s\"`\n", f.Name, f.GoType, f.JSONName))
	}
	return fmt.Sprintf(`package model

import (
	"time"

	"gorm.io/gorm"
)

// %s generated by cmd/gen crud.
type %s struct {
	ID        int64          %s
	TenantID  string         %s
%s	CreatedAt time.Time      %s
	UpdatedAt time.Time      %s
	DeletedAt gorm.DeletedAt %s
}

func (%s) TableName() string {
	return %q
}
`, modelName, modelName, "`gorm:\"primaryKey;autoIncrement\" json:\"id\"`", "`gorm:\"size:64;not null;default:default;index\" json:\"tenant_id\"`", fieldLines.String(), "`json:\"created_at\"`", "`json:\"updated_at\"`", "`gorm:\"index\" json:\"-\"`", modelName, table)
}

func daoTemplate(modelName, daoName string) string {
	return fmt.Sprintf(`package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/tenant"
)

// %s generated by cmd/gen crud.
type %s struct {
	db *gorm.DB
}

func New%s(db *gorm.DB) *%s {
	return &%s{db: db}
}

func (d *%s) Create(ctx context.Context, in *model.%s) error {
	if in != nil && in.TenantID == "" {
		in.TenantID = tenant.FromContext(ctx)
		if in.TenantID == "" {
			in.TenantID = "default"
		}
	}
	return d.db.WithContext(ctx).Create(in).Error
}

func (d *%s) Update(ctx context.Context, in *model.%s) error {
	return d.db.WithContext(ctx).Save(in).Error
}

func (d *%s) GetByID(ctx context.Context, id int64) (*model.%s, error) {
	var row model.%s
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *%s) List(ctx context.Context, offset, limit int) ([]model.%s, int64, error) {
	var total int64
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx).Model(&model.%s{}), "tenant_id").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.%s
	if err := tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").Order("id desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (d *%s) Delete(ctx context.Context, id int64) error {
	return tenant.ApplyScope(ctx, d.db.WithContext(ctx), "tenant_id").Delete(&model.%s{}, id).Error
}
`, daoName, daoName, daoName, daoName, daoName, daoName, modelName, daoName, modelName, daoName, modelName, modelName, daoName, modelName, modelName, modelName, daoName, modelName)
}

func portTemplate(modelName, serviceName string) string {
	return fmt.Sprintf(`package port

import (
	"context"

	"gin-scaffold/internal/model"
)

// %s generated by cmd/gen crud.
type %s interface {
	Create(ctx context.Context, in *model.%s) error
	Update(ctx context.Context, in *model.%s) error
	GetByID(ctx context.Context, id int64) (*model.%s, error)
	List(ctx context.Context, page, pageSize int) ([]model.%s, int64, error)
	Delete(ctx context.Context, id int64) error
}
`, serviceName, serviceName, modelName, modelName, modelName, modelName)
}

func serviceTemplate(modelName, serviceName string) string {
	return fmt.Sprintf(`package service

import (
	"context"

	"gin-scaffold/internal/model"
)

type %sRepo interface {
	Create(ctx context.Context, in *model.%s) error
	Update(ctx context.Context, in *model.%s) error
	GetByID(ctx context.Context, id int64) (*model.%s, error)
	List(ctx context.Context, offset, limit int) ([]model.%s, int64, error)
	Delete(ctx context.Context, id int64) error
}

// %s generated by cmd/gen crud.
type %s struct {
	dao %sRepo
}

func New%s(d %sRepo) *%s {
	return &%s{dao: d}
}

func (s *%s) Create(ctx context.Context, in *model.%s) error {
	return s.dao.Create(ctx, in)
}

func (s *%s) Update(ctx context.Context, in *model.%s) error {
	return s.dao.Update(ctx, in)
}

func (s *%s) GetByID(ctx context.Context, id int64) (*model.%s, error) {
	return s.dao.GetByID(ctx, id)
}

func (s *%s) List(ctx context.Context, page, pageSize int) ([]model.%s, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.dao.List(ctx, offset, pageSize)
}

func (s *%s) Delete(ctx context.Context, id int64) error {
	return s.dao.Delete(ctx, id)
}
`, serviceName, modelName, modelName, modelName, modelName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, serviceName, modelName, serviceName, modelName, serviceName, modelName, serviceName, modelName, serviceName)
}

func requestTemplate(modelName string, fields []genField) string {
	var createFields strings.Builder
	var updateFields strings.Builder
	for _, f := range fields {
		createBinding := strings.TrimSpace(f.Validate)
		if createBinding == "" && !f.Optional {
			createBinding = "required"
		}
		updateBinding := strings.TrimSpace(f.Validate)
		if updateBinding != "" {
			updateBinding = "omitempty," + updateBinding
		}
		if createBinding != "" {
			createFields.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\" binding:\"%s\"`\n", f.Name, f.GoType, f.JSONName, createBinding))
		} else {
			createFields.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", f.Name, f.GoType, f.JSONName))
		}
		if updateBinding != "" {
			updateFields.WriteString(fmt.Sprintf("\t%s *%s `json:\"%s\" binding:\"%s\"`\n", f.Name, f.GoType, f.JSONName, updateBinding))
		} else {
			updateFields.WriteString(fmt.Sprintf("\t%s *%s `json:\"%s\"`\n", f.Name, f.GoType, f.JSONName))
		}
	}
	return fmt.Sprintf(`package adminreq

// %sCreateRequest generated by cmd/gen crud.
type %sCreateRequest struct {
%s}

// %sUpdateRequest generated by cmd/gen crud.
type %sUpdateRequest struct {
%s}

// %sIDURI generated by cmd/gen crud.
type %sIDURI struct {
	ID int64 %s
}
`, modelName, modelName, createFields.String(), modelName, modelName, updateFields.String(), modelName, modelName, "`uri:\"id\" binding:\"required,min=1\"`")
}

func adminHandlerTemplate(modelName, serviceName string, fields []genField) string {
	var createAssign strings.Builder
	var updateAssign strings.Builder
	for _, f := range fields {
		createAssign.WriteString(fmt.Sprintf("\t\t%s: req.%s,\n", f.Name, f.Name))
		updateAssign.WriteString(fmt.Sprintf("\tif req.%s != nil {\n\t\tin.%s = *req.%s\n\t}\n", f.Name, f.Name, f.Name))
	}
	tpl := `package adminhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	adminreq "gin-scaffold/api/request/admin"
	"gin-scaffold/api/response"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/internal/service/port"
)

// {{Model}}Handler generated by cmd/gen crud.
type {{Model}}Handler struct {
	svc port.{{Service}}
}

func New{{Model}}Handler(s port.{{Service}}) *{{Model}}Handler {
	return &{{Model}}Handler{svc: s}
}

func (h *{{Model}}Handler) List(c *gin.Context) {
	var q adminreq.PageQuery
	_ = c.ShouldBindQuery(&q)
	rows, total, err := h.svc.List(c.Request.Context(), q.Page, q.PageSize)
	if err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"total": total, "list": rows})
}

func (h *{{Model}}Handler) Get(c *gin.Context) {
	var uri adminreq.{{Model}}IDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	row, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		response.FailHTTP(c, http.StatusNotFound, errcode.NotFound, errcode.KeyInvalidParam, err.Error())
		return
	}
	response.OK(c, row)
}

func (h *{{Model}}Handler) Create(c *gin.Context) {
	var req adminreq.{{Model}}CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	in := &model.{{Model}}{
{{CreateAssign}}	}
	if err := h.svc.Create(c.Request.Context(), in); err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, in)
}

func (h *{{Model}}Handler) Update(c *gin.Context) {
	var uri adminreq.{{Model}}IDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	var req adminreq.{{Model}}UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	in := &model.{{Model}}{ID: uri.ID}
{{UpdateAssign}}	if err := h.svc.Update(c.Request.Context(), in); err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, in)
}

func (h *{{Model}}Handler) Delete(c *gin.Context) {
	var uri adminreq.{{Model}}IDURI
	if err := c.ShouldBindUri(&uri); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uri.ID); err != nil {
		response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
`
	replacer := strings.NewReplacer(
		"{{Model}}", modelName,
		"{{Service}}", serviceName,
		"{{CreateAssign}}", createAssign.String(),
		"{{UpdateAssign}}", updateAssign.String(),
	)
	return replacer.Replace(tpl)
}

func adminRouteTemplate(modelName string) string {
	lower := strings.ToLower(modelName)
	tpl := `package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	"gin-scaffold/middleware"
)

// registerAdmin{{Model}}Routes generated by cmd/gen crud.
func registerAdmin{{Model}}Routes(admin *gin.RouterGroup, h *adminhandler.{{Model}}Handler) {
	admin.GET("/{{resource}}s", middleware.RequirePermission("{{perm}}:read"), h.List)
	admin.GET("/{{resource}}s/:id", middleware.RequirePermission("{{perm}}:read"), h.Get)
	admin.POST("/{{resource}}s", middleware.RequirePermission("{{perm}}:write"), h.Create)
	admin.PUT("/{{resource}}s/:id", middleware.RequirePermission("{{perm}}:write"), h.Update)
	admin.DELETE("/{{resource}}s/:id", middleware.RequirePermission("{{perm}}:write"), h.Delete)
}
`
	replacer := strings.NewReplacer(
		"{{Model}}", modelName,
		"{{resource}}", lower,
		"{{perm}}", lower,
	)
	return replacer.Replace(tpl)
}

func adminRegisterTemplate() string {
	return `package adminroutes

import (
	"github.com/gin-gonic/gin"

	adminhandler "gin-scaffold/api/handler/admin"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/middleware"
)

func Register(r *gin.Engine, jwtMgr *jwtpkg.Manager, user *adminhandler.UserHandler, menu *adminhandler.MenuHandler, ops *adminhandler.OpsHandler, task *adminhandler.TaskHandler, sys *adminhandler.SystemSettingHandler, queue *adminhandler.TaskQueueHandler, generatedAnnouncement *adminhandler.AnnouncementHandler) {
	if jwtMgr == nil {
		return
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.JWTAuth(jwtMgr))
	admin.Use(middleware.RequireRoles("admin"))
	registerAdminUserRoutes(admin, user)
	registerAdminMenuRoutes(admin, menu)
	registerAdminOpsRoutes(admin, ops)
	registerAdminTaskRoutes(admin, task)
	registerAdminTaskQueueRoutes(admin, queue)
	registerAdminSystemSettingRoutes(admin, sys)
	registerAdminAnnouncementRoutes(admin, generatedAnnouncement)
}
`
}

func schemaUpTemplate(table string, fields []genField) string {
	var cols strings.Builder
	for _, f := range fields {
		def := formatSQLDefault(f)
		if def != "" {
			cols.WriteString(fmt.Sprintf("  `%s` %s NOT NULL DEFAULT %s,\n", f.JSONName, f.SQLType, def))
		} else {
			cols.WriteString(fmt.Sprintf("  `%s` %s NOT NULL,\n", f.JSONName, f.SQLType))
		}
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n  `id` BIGINT NOT NULL AUTO_INCREMENT,\n  `tenant_id` VARCHAR(64) NOT NULL DEFAULT 'default',\n%s  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),\n  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),\n  `deleted_at` DATETIME(3) NULL,\n  PRIMARY KEY (`id`),\n  KEY `idx_%s_tenant_id` (`tenant_id`),\n  KEY `idx_%s_deleted_at` (`deleted_at`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;\n", table, cols.String(), table, table)
}

func formatSQLDefault(f genField) string {
	v := strings.TrimSpace(f.Default)
	if v == "" {
		return ""
	}
	switch f.GoType {
	case "string":
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case "bool":
		l := strings.ToLower(v)
		if l == "true" || l == "1" {
			return "1"
		}
		return "0"
	default:
		return v
	}
}

func schemaDownTemplate(table string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %q;\n", table)
}

func seedPermUpTemplate(prefix string) string {
	return fmt.Sprintf(`INSERT IGNORE INTO role_permissions (tenant_id, role, permission, created_at, updated_at) VALUES
  ('default', 'admin', '%s:read', NOW(), NOW()),
  ('default', 'admin', '%s:write', NOW(), NOW());
`, prefix, prefix)
}

func seedPermDownTemplate(prefix string) string {
	return fmt.Sprintf("DELETE FROM role_permissions WHERE tenant_id = 'default' AND role = 'admin' AND permission IN ('%s:read', '%s:write');\n", prefix, prefix)
}
