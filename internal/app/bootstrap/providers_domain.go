package bootstrap

import (
	"github.com/hibiken/asynq"
	"gorm.io/gorm"

	"gin-scaffold/internal/api/handler"
	adminhandler "gin-scaffold/internal/api/handler/admin"
	clienthandler "gin-scaffold/internal/api/handler/client"
	"gin-scaffold/internal/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	"gin-scaffold/internal/middleware"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	websocketpkg "gin-scaffold/internal/pkg/websocket"
	"gin-scaffold/internal/service"
	"gin-scaffold/internal/service/authz"
)

type daoProviders struct {
	user         *dao.UserDAO
	menu         *dao.MenuDAO
	task         *dao.ScheduledTaskDAO
	audit        *dao.AuditLogDAO
	outbox       *dao.OutboxDAO
	system       *dao.SystemSettingDAO
	announcement *dao.AnnouncementDAO
	authz        *dao.AuthzDAO
}

type serviceProviders struct {
	user         *service.UserService
	menu         *service.MenuService
	task         *service.ScheduledTaskService
	system       *service.SystemSettingService
	announcement *service.AnnouncementService
	ws           *service.WSService
	sse          *service.SSEService
}

type handlerProviders struct {
	base              *handler.BaseHandler
	clientUser        *clienthandler.UserHandler
	clientFile        *clienthandler.FileHandler
	adminUser         *adminhandler.UserHandler
	adminMenu         *adminhandler.MenuHandler
	adminOps          *adminhandler.OpsHandler
	adminTask         *adminhandler.TaskHandler
	adminSystem       *adminhandler.SystemSettingHandler
	adminQueue        *adminhandler.TaskQueueHandler
	adminAnnouncement *adminhandler.AnnouncementHandler
	ws                *handler.WSHandler
	sse               *handler.SSEHandler
}

func provideDAO(gdb *gorm.DB) daoProviders {
	return daoProviders{
		user:         dao.NewUserDAO(gdb),
		menu:         dao.NewMenuDAO(gdb),
		task:         dao.NewScheduledTaskDAO(gdb),
		audit:        dao.NewAuditLogDAO(gdb),
		outbox:       dao.NewOutboxDAO(gdb),
		system:       dao.NewSystemSettingDAO(gdb),
		announcement: dao.NewAnnouncementDAO(gdb),
		authz:        dao.NewAuthzDAO(gdb),
	}
}

func provideAuthzInfra(cfg *config.App, daos daoProviders) {
	middleware.SetPermissionChecker(authz.NewDBPermissionChecker(daos.authz, cfg.RBAC.SuperAdminUserID))
	middleware.SetSuperAdminUserID(cfg.RBAC.SuperAdminUserID)
}

func provideServices(cfg *config.App, gdb *gorm.DB, q *job.Client, jm *jwtpkg.Manager, daos daoProviders) serviceProviders {
	userSvc := service.NewUserService(daos.user, q, jm, cfg.RBAC.SuperAdminUserID, gdb, daos.outbox, cfg.Outbox)
	menuSvc := service.NewMenuService(daos.menu)
	taskSvc := service.NewScheduledTaskService(daos.task, cfg.Scheduler)
	systemSvc := service.NewSystemSettingService(daos.system)
	announcementSvc := service.NewAnnouncementService(daos.announcement)
	hub := websocketpkg.NewHub()
	return serviceProviders{
		user:         userSvc,
		menu:         menuSvc,
		task:         taskSvc,
		system:       systemSvc,
		announcement: announcementSvc,
		ws:           service.NewWSService(hub),
		sse:          service.NewSSEService(),
	}
}

func provideHandlers(cfg *config.App, gdb *gorm.DB, q *job.Client, inspector *asynq.Inspector, svcs serviceProviders, daos daoProviders) handlerProviders {
	return handlerProviders{
		base:              &handler.BaseHandler{DB: gdb, Storage: &cfg.Storage},
		clientUser:        clienthandler.NewUserHandler(svcs.user),
		clientFile:        clienthandler.NewFileHandler(&cfg.Storage),
		adminUser:         adminhandler.NewUserHandler(svcs.user, q),
		adminMenu:         adminhandler.NewMenuHandler(svcs.menu),
		adminOps:          adminhandler.NewOpsHandler(daos.audit, q),
		adminTask:         adminhandler.NewTaskHandler(svcs.task),
		adminSystem:       adminhandler.NewSystemSettingHandler(svcs.system),
		adminQueue:        adminhandler.NewTaskQueueHandler(inspector),
		adminAnnouncement: adminhandler.NewAnnouncementHandler(svcs.announcement),
		ws:                handler.NewWSHandler(svcs.ws, middleware.WebSocketCheckOrigin(cfg.CORS.AllowOrigins)),
		sse:               handler.NewSSEHandler(svcs.sse),
	}
}
