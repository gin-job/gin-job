package main

import (
	"github.com/gin-job/gin-job/config"
	"github.com/gin-job/gin-job/examples/simple/jobs"
	"github.com/gin-job/gin-job/job"
	"github.com/gin-job/gin-job/model"
	"github.com/gin-job/gin-job/router"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.DefaultConfig()
	// init logger
	zapLogger, _ := zap.NewProduction()
	// init db
	dsn := "root:gin-job@tcp(localhost:3306)/gin_job?charset=utf8mb4&parseTime=True&loc=Local"
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zapLogger.Fatal("数据库连接失败", zap.Error(err))
	}

	// db auto migrate
	if err := gormDB.AutoMigrate(
		&model.SysJobScheduleModel{},
		&model.SysJobInstanceModel{},
	); err != nil {
		zapLogger.Fatal("数据库迁移失败", zap.Error(err))
	}

	// register job
	jobList := []job.Job{
		&jobs.ExampleJob{},
	}

	// init router
	r := router.NewGinJobRouter(zapLogger, gormDB, cfg, jobList)
	r.Start()
}
