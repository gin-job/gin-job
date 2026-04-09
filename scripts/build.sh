#!/bin/bash
# 读取version文件
VERSION=$(cat version)
# 构建镜像
docker build -f examples/simple/Dockerfile -t ghcr.io/gin-job/simple:latest .
# 标记镜像版本
docker tag ghcr.io/gin-job/simple:latest ghcr.io/gin-job/simple:$VERSION
