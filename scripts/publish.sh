# 读取version文件
VERSION=$(cat version)
# 推送镜像
docker push ghcr.io/gin-job/simple:$VERSION
# 推送最新版本
docker push ghcr.io/gin-job/simple:latest
