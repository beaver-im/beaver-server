# 通用运行镜像：二进制由 build/build.sh 通过 BIN 参数指定（来自 beaver_server 构建阶段）
# 配置文件由各服务 docker-compose.yaml 挂载到 /etc/config.yaml
FROM beaver_server AS builder
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

ARG BIN
COPY --from=builder /build/${BIN} ./app

CMD ["./app", "-f", "/etc/config.yaml"]
