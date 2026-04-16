package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

type crudOptions struct {
	module string
	table  string
	force  bool
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
	crudCmd.Flags().BoolVar(&opt.force, "force", false, "overwrite existing files")
	_ = crudCmd.MarkFlagRequired("module")
	rootCmd.AddCommand(crudCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/gen crud --module <name> [--table <table_name>] [--force]")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("  go run ./cmd/gen crud --module order --table orders --force")
}

func runCRUD(opt crudOptions) error {
	opt.module = strings.TrimSpace(opt.module)
	if opt.module == "" {
		return errors.New("module is required")
	}
	if opt.table == "" {
		opt.table = toSnake(opt.module) + "s"
	}

	modelName := toPascal(opt.module)
	fieldName := lowerFirst(modelName)
	moduleSnake := toSnake(opt.module)
	daoName := modelName + "DAO"
	serviceName := modelName + "Service"

	files := map[string]string{
		filepath.Join("internal", "model", moduleSnake+".go"):                   modelTemplate(modelName, opt.table),
		filepath.Join("internal", "dao", moduleSnake+"_dao.go"):                 daoTemplate(modelName, daoName),
		filepath.Join("internal", "service", "port", moduleSnake+"_service.go"): portTemplate(modelName, serviceName),
		filepath.Join("internal", "service", moduleSnake+"_service.go"):         serviceTemplate(modelName, serviceName),
		filepath.Join("api", "request", "admin", moduleSnake+"_request.go"):     requestTemplate(modelName),
		filepath.Join("api", "handler", "admin", moduleSnake+"_handler.go"):     adminHandlerTemplate(modelName, serviceName, fieldName),
		filepath.Join("routes", "admin_"+moduleSnake+"_router.go"):              adminRouteTemplate(modelName),
	}

	for p, content := range files {
		if err := writeFile(p, content, opt.force); err != nil {
			return err
		}
	}

	if err := wireGeneratedCRUD(moduleSnake, modelName, daoName, serviceName); err != nil {
		return err
	}

	fmt.Println("CRUD scaffold generated:")
	for p := range files {
		fmt.Println(" -", p)
	}
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1) fill request/response field mapping in generated handler")
	fmt.Println("  2) add migration SQL for table:", opt.table)
	fmt.Println("  3) add permissions and role mapping for module:", strings.ToLower(modelName))

	return nil
}

func writeFile(relPath, content string, force bool) error {
	abs := filepath.Clean(relPath)
	if _, err := os.Stat(abs); err == nil && !force {
		return fmt.Errorf("file exists: %s (use --force to overwrite)", relPath)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
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
			"\tAdminOps   *adminhandler.OpsHandler\n",
			"\tAdminOps   *adminhandler.OpsHandler\n"+fieldLine,
			1,
		)
	}

	arg := fmt.Sprintf("opts.Admin%s", modelName)
	if !strings.Contains(text, arg) {
		text = strings.Replace(
			text,
			"opts.AdminOps, opts.WS, opts.SSE",
			fmt.Sprintf("opts.AdminOps, %s, opts.WS, opts.SSE", arg),
			1,
		)
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
			"\tadminOps *adminhandler.OpsHandler,\n",
			"\tadminOps *adminhandler.OpsHandler,\n"+paramLine,
			1,
		)
	}

	arg := fmt.Sprintf("admin%s", modelName)
	if !strings.Contains(text, "registerAdminRoutes(r, jwtMgr, adminUser, adminMenu, adminOps, "+arg+")") {
		text = strings.Replace(
			text,
			"registerAdminRoutes(r, jwtMgr, adminUser, adminMenu, adminOps)",
			"registerAdminRoutes(r, jwtMgr, adminUser, adminMenu, adminOps, "+arg+")",
			1,
		)
	}

	return os.WriteFile(path, []byte(text), 0o644)
}

func wireAdminRouter(modelName string) error {
	path := filepath.Join("routes", "admin_router.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(data)

	param := fmt.Sprintf("generated%s *adminhandler.%sHandler", modelName, modelName)
	if !strings.Contains(text, param) {
		text = strings.Replace(
			text,
			"ops *adminhandler.OpsHandler",
			"ops *adminhandler.OpsHandler, "+param,
			1,
		)
	}

	callLine := fmt.Sprintf("\tregisterAdmin%sRoutes(admin, generated%s)\n", modelName, modelName)
	if !strings.Contains(text, callLine) {
		text = strings.Replace(text, "\tadmin.GET(\"/dbping\", middleware.RequirePermission(\"db:ping\"), ops.DBPing)\n", "\tadmin.GET(\"/dbping\", middleware.RequirePermission(\"db:ping\"), ops.DBPing)\n"+callLine, 1)
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
		text = strings.Replace(text, "\tauthzDAO := dao.NewAuthzDAO(gdb)\n", "\tauthzDAO := dao.NewAuthzDAO(gdb)\n"+daoLine, 1)
	}

	svcLine := fmt.Sprintf("\t%sSvc := service.New%s(%sDAO)\n", lowerFirst(modelName), serviceName, lowerFirst(modelName))
	if !strings.Contains(text, svcLine) {
		text = strings.Replace(text, "\tmenuSvc := service.NewMenuService(menuDAO)\n", "\tmenuSvc := service.NewMenuService(menuDAO)\n"+svcLine, 1)
	}

	handlerLine := fmt.Sprintf("\tadmin%sH := adminhandler.New%sHandler(%sSvc)\n", modelName, modelName, lowerFirst(modelName))
	if !strings.Contains(text, handlerLine) {
		text = strings.Replace(text, "\tadminOpsH := adminhandler.NewOpsHandler()\n", "\tadminOpsH := adminhandler.NewOpsHandler()\n"+handlerLine, 1)
	}

	optLine := fmt.Sprintf("\t\tAdmin%s:  admin%sH,\n", modelName, modelName)
	if !strings.Contains(text, optLine) {
		text = strings.Replace(text, "\t\tAdminOps:   adminOpsH,\n", "\t\tAdminOps:   adminOpsH,\n"+optLine, 1)
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

func modelTemplate(modelName, table string) string {
	return fmt.Sprintf(`package model

import (
	"time"

	"gorm.io/gorm"
)

// %s generated by cmd/gen crud.
type %s struct {
	ID        int64          %s
	CreatedAt time.Time      %s
	UpdatedAt time.Time      %s
	DeletedAt gorm.DeletedAt %s
}

func (%s) TableName() string {
	return %q
}
`, modelName, modelName, "`gorm:\"primaryKey;autoIncrement\" json:\"id\"`", "`json:\"created_at\"`", "`json:\"updated_at\"`", "`gorm:\"index\" json:\"-\"`", modelName, table)
}

func daoTemplate(modelName, daoName string) string {
	return fmt.Sprintf(`package dao

import (
	"context"

	"gorm.io/gorm"

	"gin-scaffold/internal/model"
)

// %s generated by cmd/gen crud.
type %s struct {
	db *gorm.DB
}

func New%s(db *gorm.DB) *%s {
	return &%s{db: db}
}

func (d *%s) Create(ctx context.Context, in *model.%s) error {
	return d.db.WithContext(ctx).Create(in).Error
}

func (d *%s) Update(ctx context.Context, in *model.%s) error {
	return d.db.WithContext(ctx).Save(in).Error
}

func (d *%s) GetByID(ctx context.Context, id int64) (*model.%s, error) {
	var row model.%s
	if err := d.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *%s) List(ctx context.Context, offset, limit int) ([]model.%s, int64, error) {
	var total int64
	if err := d.db.WithContext(ctx).Model(&model.%s{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.%s
	if err := d.db.WithContext(ctx).Order("id desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (d *%s) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&model.%s{}, id).Error
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

func requestTemplate(modelName string) string {
	return fmt.Sprintf(`package adminreq

// %sCreateRequest generated by cmd/gen crud.
type %sCreateRequest struct{}

// %sUpdateRequest generated by cmd/gen crud.
type %sUpdateRequest struct{}

// %sIDURI generated by cmd/gen crud.
type %sIDURI struct {
	ID int64 %s
}
`, modelName, modelName, modelName, modelName, modelName, modelName, "`uri:\"id\" binding:\"required,min=1\"`")
}

func adminHandlerTemplate(modelName, serviceName, fieldName string) string {
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
	{{Field}}, err := h.svc.GetByID(c.Request.Context(), uri.ID)
	if err != nil {
		response.FailHTTP(c, http.StatusNotFound, errcode.NotFound, errcode.KeyInvalidParam, err.Error())
		return
	}
	response.OK(c, {{Field}})
}

func (h *{{Model}}Handler) Create(c *gin.Context) {
	var req adminreq.{{Model}}CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
		return
	}
	// TODO: map req to model.{{Model}} fields.
	in := &model.{{Model}}{}
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
	// TODO: map req to model.{{Model}} fields.
	in := &model.{{Model}}{ID: uri.ID}
	if err := h.svc.Update(c.Request.Context(), in); err != nil {
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
		"{{Field}}", fieldName,
	)
	return replacer.Replace(tpl)
}

func adminRouteTemplate(modelName string) string {
	lower := strings.ToLower(modelName)
	tpl := `package routes

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
