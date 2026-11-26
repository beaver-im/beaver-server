#!/bin/bash

REGISTRY_URL="wsrh8888"
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="$(tr -d '[:space:]' < "$ROOT_DIR/VERSION")"
PUBLISH_PROBE_MODULE="gateway_api"

# 检查当前版本是否已发布到镜像仓库
assert_version_not_published() {
    local image="$REGISTRY_URL/$PUBLISH_PROBE_MODULE:$VERSION"
    echo "🔍 检查版本 $VERSION 是否已发布..."
    if docker manifest inspect "$image" >/dev/null 2>&1; then
        echo "❌ 版本 $VERSION 已发布（$image 已存在于镜像仓库）"
        echo "   请先更新 $ROOT_DIR/VERSION 后再执行构建发布"
        exit 1
    fi
    echo "✅ 版本 $VERSION 尚未发布，开始构建"
}

assert_version_not_published

# 定义模块列表（与 app/、deploy/、Dockerfile 保持一致，服务合并后共 16 个）
modules=(
    # ==================== RPC 服务 ====================
    "account_rpc"
    "asset_rpc"
    "message_rpc"
    "notification_rpc"
    "platform_rpc"
    "relation_rpc"

    # ==================== API 服务 ====================
    "account_api"
    "asset_api"
    "datasync_api"
    "gateway_api"
    "message_api"
    "notification_api"
    "platform_api"
    "push_api"
    "relation_api"
    "ws_api"
)

# 在仓库根目录执行 docker build（不依赖调用方当前工作目录）
docker build -t beaver_server -f "$ROOT_DIR/Dockerfile" "$ROOT_DIR"

# 记录失败模块，避免某个模块推送失败但仍报告整体成功
FAIL_FLAG="$(mktemp)"
trap 'rm -f "$FAIL_FLAG"' EXIT

# 并发构建函数
build_module() {
    local module_name="$1"
    echo "🚀 开始构建: $module_name"
    if bash "$ROOT_DIR/build/build.sh" "$module_name"; then
        echo "✅ 打包镜像成功: $module_name"
    else
        echo "❌ 打包镜像失败: $module_name" >&2
        echo "$module_name" >> "$FAIL_FLAG"
    fi
    echo "----------------------------------------"
}

# 并发构建，限制最大并发数为4
max_jobs=4
current_jobs=0

for module_name in "${modules[@]}"; do
    # 如果当前并发数达到上限，等待一个任务完成
    while [ $current_jobs -ge $max_jobs ]; do
        wait -n
        current_jobs=$((current_jobs - 1))
    done
    
    # 启动后台任务
    build_module "$module_name" &
    current_jobs=$((current_jobs + 1))
done

# 等待所有后台任务完成
wait

if [ -s "$FAIL_FLAG" ]; then
    echo "❌ 以下模块构建/推送失败：$(tr '\n' ' ' < "$FAIL_FLAG")" >&2
    exit 1
fi

echo "🎉 所有模块构建并推送完成！"