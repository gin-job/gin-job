package router

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-job/gin-job/config"
	"github.com/gin-job/gin-job/handler"
	"github.com/gin-job/gin-job/job"
	"github.com/gin-job/gin-job/scheduler"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewGinJobRouter(
	logger *zap.Logger,
	gormDB *gorm.DB,
	config *config.GinJobConfig,
	jobList []job.Job,
) *GinJobRouter {
	sch := scheduler.New(logger, gormDB)
	return &GinJobRouter{
		logger:  logger,
		sch:     sch,
		gormDB:  gormDB,
		config:  config,
		jobList: jobList,
	}
}

type GinJobRouter struct {
	logger  *zap.Logger
	sch     *scheduler.Scheduler
	gormDB  *gorm.DB
	config  *config.GinJobConfig
	jobList []job.Job
}

func (g *GinJobRouter) Start() {
	logger := g.logger
	gormDB := g.gormDB
	jobList := g.jobList
	config := g.config
	sch := g.sch

	// init scheduler
	if err := sch.SyncFromDB(); err != nil {
		logger.Error("同步数据库任务失败", zap.Error(err))
	}
	// register jobs
	for _, item := range jobList {
		if err := job.Register(item); err != nil {
			logger.Error("register job failed", zap.Error(err))
		}
	}
	// start scheduler
	sch.Start()

	// init router
	r := gin.Default()
	// TODO：需要代码审查
	r.LoadHTMLGlob(config.TemplatePath)

	// register routes
	if sch != nil && gormDB != nil {
		h := handler.NewJobHandler(sch, logger, gormDB)
		h.RegisterRoutes(r)
		handler.NewUIRoutes(&config.Auth).RegisterRoutes(r)
	}

	// start router
	go func() {
		if err := r.Run(config.Port); err != nil {
			logger.Error("Job HTTP Server start failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Job HTTP Server received exit signal, prepare to close")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sch.Stop(ctx)
	logger.Info("Job Scheduler stopped successfully")
}
