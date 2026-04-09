# gin-job

A task scheduling management tool based on the Gin framework, with a user-friendly UI interface for managing scheduled tasks.

## Features

- Based on the Gin framework for task scheduling management
- With UI interface for easy task management
- Supports task definition, execution, pause, resume, and deletion operations
- Supports task history log query

## Dependencies

- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [Zap](https://github.com/uber-go/zap) - Logging library
- [Cron](https://github.com/robfig/cron) - Task scheduling library

## Quick Start

First, initialize the local database:

```bash
docker run -it -d -p 3306:3306 
-e MYSQL_ROOT_PASSWORD=gin-job \
-e MYSQL_DATABASE=gin_job \
mysql:8.0.42
```

Run the gin-job example application:

`go run examples/simple/main.go`

Open the UI interface in your browser:

`http://localhost:8080/ui/login`

## How to Contribute

- Send an email to [project maintainer](mailto:liuhong@neme.ai) or submit a Pull Request to the project repository to request to join the contributor list

## Detailed Documentation

For more detailed documentation, please refer to [GinJob.md](docs/GinJob.md)