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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewGinJobRouter(cfg *config.GinJobConfig) *GinJobRouter {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	logger, _ := zap.NewProduction()
	gormDB, err := gorm.Open(mysql.Open(cfg.Gorm.DSN), cfg.Gorm.Config)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}
	sch := scheduler.New(logger, gormDB)

	return &GinJobRouter{
		logger:    logger,
		scheduler: sch,
		gormDB:    gormDB,
		config:    cfg,
		jobList:   []job.Job{},
	}
}

type GinJobRouter struct {
	logger    *zap.Logger
	scheduler *scheduler.Scheduler
	gormDB    *gorm.DB
	config    *config.GinJobConfig
	jobList   []job.Job
}

func (g *GinJobRouter) Start() {
	// register jobs
	for _, item := range g.jobList {
		if err := job.Register(item); err != nil {
			g.logger.Error("register job failed", zap.Error(err))
		}
	}
	// init scheduler
	if err := g.scheduler.SyncFromDB(); err != nil {
		g.logger.Error("sync jobs from db failed", zap.Error(err))
	}
	// start scheduler
	g.scheduler.Start()

	// init router
	r := gin.Default()

	if g.config.TemplatePath != "" {
		r.LoadHTMLGlob(g.config.TemplatePath)
	} else {
		r.LoadHTMLGlob("templates/*")
	}

	// register routes
	if g.scheduler != nil && g.gormDB != nil {
		h := handler.NewJobHandler(g.scheduler, g.logger, g.gormDB)
		h.RegisterRoutes(r)
		handler.NewUIRoutes(&g.config.Auth).RegisterRoutes(r)
	}

	// start router
	go func() {
		if err := r.Run(g.config.Port); err != nil {
			g.logger.Error("Job HTTP Server start failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	g.logger.Info("Job HTTP Server received exit signal, prepare to close")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	g.scheduler.Stop(ctx)
	g.logger.Info("Job Scheduler stopped successfully")
}

func (g *GinJobRouter) SetJobList(jobList []job.Job) {
	g.jobList = jobList
}
