# Builder stage
FROM ubuntu:22.04 as builder
ARG GO_VERSION=1.26.2

# Add Alibaba mirror source
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list
# Update package list and install wget
RUN apt-get update && apt-get install -y wget

# Install Go
RUN wget -q https://mirrors.aliyun.com/golang/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/local/bin/go && \
    go version

RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# Build gin-job
WORKDIR /app
# Copy the entire project structure to access templates
COPY . .
RUN cd examples/simple && \
    go mod download -x && \
    go build -o gin-job main.go

# Final stage
FROM ubuntu/mysql:8.0-22.04_beta
ENV MYSQL_ROOT_PASSWORD=gin-
ENV TEMPLATE_PATH=/app/templates/*

# Add Alibaba mirror source
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
# RUN sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list
RUN apt-get update
# Run MySQL entrypoint
RUN docker-entrypoint.sh

WORKDIR /app
# Copy gin-job from builder stage
COPY --from=builder /app/examples/simple/gin-job /app/gin-job
# Copy templates from builder stage
COPY --from=builder /app/templates /app/templates/

EXPOSE 8080
EXPOSE 3306
EXPOSE 33060

ENTRYPOINT ["gin-job"]