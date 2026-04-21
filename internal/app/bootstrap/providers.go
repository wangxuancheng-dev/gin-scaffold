package bootstrap

import (
	"gorm.io/gorm"

	"gin-scaffold/internal/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/routes"
	"gin-scaffold/internal/service"
	"github.com/hibiken/asynq"
)

type serverProviders struct {
	taskSvc   *service.ScheduledTaskService
	outboxDAO *dao.OutboxDAO
	routeOpts routes.Options
}

type workerProviders struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func buildServerProviders(cfg *config.App, gdb *gorm.DB, q *job.Client, inspector *asynq.Inspector) *serverProviders {
	daos := provideDAO(gdb)
	jm := jwtpkg.NewManager(&cfg.JWT)
	provideAuthzInfra(cfg, daos)
	svcs := provideServices(cfg, gdb, q, jm, daos)
	handlers := provideHandlers(cfg, gdb, q, inspector, svcs, daos)

	return &serverProviders{
		taskSvc:   svcs.task,
		outboxDAO: daos.outbox,
		routeOpts: routes.Options{
			Cfg:               cfg,
			JWT:               jm,
			Base:              handlers.base,
			ClientUser:        handlers.clientUser,
			ClientFile:        handlers.clientFile,
			AdminUser:         handlers.adminUser,
			AdminMenu:         handlers.adminMenu,
			AdminOps:          handlers.adminOps,
			AdminTask:         handlers.adminTask,
			AdminSys:          handlers.adminSystem,
			AdminQueue:        handlers.adminQueue,
			AdminAnnouncement: handlers.adminAnnouncement,
			WS:                handlers.ws,
			SSE:               handlers.sse,
			TraceOn:           cfg.Trace.Enabled,
			Limiter:           buildLimiterProvider(cfg),
		},
	}
}
