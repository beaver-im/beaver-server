#!/bin/bash

# 获取当前脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 遍历所有含 docker-compose.yaml 的目录并启动。
# 每个业务目录的 compose 内 RPC/API 用 depends_on 约束启动顺序，无需在此按后缀分批。
for dir in "$SCRIPT_DIR"/*; do
    [ -d "$dir" ] || continue
    compose="$dir/docker-compose.yaml"
    [ -f "$compose" ] || continue

    echo "=== Starting $(basename "$dir") ==="
    cd "$dir" || { echo "Failed to navigate to directory: $dir"; continue; }

    docker-compose down
    docker-compose rm -f
    docker-compose pull
    docker-compose build
    docker-compose up -d

    echo "=== $(basename "$dir") started ==="
    sleep 2
done

# 清理未被使用的镜像
echo "Pruning unused Docker images"
docker image prune -f

echo "All services started."
