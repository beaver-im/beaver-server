#!/bin/bash

# 配置变量
REGISTRY_URL="wsrh8888"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="$(tr -d '[:space:]' < "$ROOT_DIR/VERSION")"

# 函数：获取Docker镜像的Id
get_image_id() {
    docker images -q $1 | head -n1
}

# 检查是否传入了参数
if [ -z "$1" ];then
    echo "Usage: $0 module_name"
    exit 1
fi

# 获取参数
module_name="$1"

# 二进制在 beaver_server 构建阶段的路径：如 account_api/account、asset_rpc/assetrpc
declare -A BIN_PATHS=(
    [account_api]="account_api/account"
    [account_rpc]="account_rpc/accountrpc"
    [asset_api]="asset_api/asset"
    [asset_rpc]="asset_rpc/assetrpc"
    [datasync_api]="datasync_api/datasync"
    [gateway_api]="gateway_api/gateway"
    [message_api]="message_api/message"
    [message_rpc]="message_rpc/messagerpc"
    [notification_api]="notification_api/notification"
    [notification_rpc]="notification_rpc/notificationrpc"
    [platform_api]="platform_api/platform"
    [platform_rpc]="platform_rpc/platformrpc"
    [relation_api]="relation_api/relation"
    [relation_rpc]="relation_rpc/relationrpc"
    [ws_api]="ws_api/ws"
)
bin_path="${BIN_PATHS[$module_name]}"
if [ -z "$bin_path" ]; then
    echo "Error: Unknown module $module_name (no binary path mapping)"
    exit 1
fi

# 在 deploy 根目录用通用 Dockerfile 构建（服务配置由各 compose 挂载，不进镜像）
cd "$ROOT_DIR/deploy" || {
    echo "Error: Directory deploy not found: $ROOT_DIR/deploy"
    exit 1
}

# 构建镜像
echo "Building image: $module_name (bin: $bin_path)"
docker build --build-arg "BIN=$bin_path" -t "$module_name" -f Dockerfile .
if [ $? -ne 0 ]; then
    echo "Error: Failed to build the image"
    exit 1
fi

# 获取镜像Id
image_id=$(get_image_id "$module_name")

# 检查获取到的image_id是否为空
if [ -z "$image_id" ];then
    echo "Error: Failed to retrieve the image Id"
    exit 1
fi

# 打标签并推送镜像
docker tag "$image_id" "$REGISTRY_URL/$module_name:$VERSION"
if [ $? -ne 0 ]; then
    echo "Error: Failed to tag the image"
    exit 1
fi

# 推送镜像，带重试机制
max_retries=3
retry_count=0

while [ $retry_count -lt $max_retries ]; do
    echo "Pushing image to registry (attempt $((retry_count + 1))/$max_retries)..."
    docker push "$REGISTRY_URL/$module_name:$VERSION"
    
    if [ $? -eq 0 ]; then
        echo "Successfully pushed the image"
        break
    else
        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            echo "Push failed, retrying in 5 seconds..."
            sleep 5
        else
            echo "Error: Failed to push the image to registry after $max_retries attempts"
            exit 1
        fi
    fi
done

echo "Successfully built, tagged, and pushed the image: $module_name:$VERSION"
