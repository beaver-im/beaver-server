#!/bin/bash

# 获取当前脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 停止所有含 docker-compose.yaml 的目录。
# 各目录命名已统一，业务目录的 compose 内 RPC/API 一起停，无需按后缀分类。
for dir in "$SCRIPT_DIR"/*; do
    [ -d "$dir" ] || continue
    [ -f "$dir/docker-compose.yaml" ] || continue
    name="$(basename "$dir")"

    echo "=== Stopping $name ==="
    cd "$dir" || { echo "Failed to navigate to directory: $dir"; continue; }
    docker-compose down
    echo "=== $name stopped ==="
    sleep 1
done

# 清理停止的容器（服务均用 host 网络，无需清理 docker 网络）
echo "Cleaning up stopped containers"
docker container prune -f

echo "All services stopped and cleaned up."
