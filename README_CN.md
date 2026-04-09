# gin-job

基于 Gin 框架的定时任务管理工具，带有用户友好的 UI 界面，方便管理定时任务。

## 功能介绍

- 基于 gin 框架的定时任务管理工具
- 带UI界面，方便管理定时任务
- 支持定时任务的定义、执行、暂停、恢复、删除等操作
- 支持任务历史日志查询

## 依赖项

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [Zap](https://github.com/uber-go/zap) - 日志库
- [Cron](https://github.com/robfig/cron) - 任务调度库

## 快速开始

先初始化本地数据库：

```bash
docker run -it -d -p 3306:3306 \ 
-e MYSQL_ROOT_PASSWORD=gin-job \
-e MYSQL_DATABASE=gin_job \
mysql:8.0.42
```

运行 gin-job 示例应用：

`go run examples/simple/main.go`

打开浏览器访问 UI 界面：

`http://localhost:8080/ui/login`

## 如何参与贡献

- 发送邮件给[项目负责人](mailto:liuhong@neme.ai)或在项目仓库中提交Pull Request请求，请求加入贡献者列表

## 详细文档

更多详细文档，请参考 [GinJob_CN.md](docs/GinJob_CN.md)