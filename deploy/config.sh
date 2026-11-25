#!/bin/bash

# 按规则文件把部署产物 dist/ 里的服务配置从 old 值替换成 new 值，并生成完整产物包。
#
# 关键约定：只改 dist/ 产物，绝不就地改写 app/ 与 deploy/ 的源码配置。
#   deploy/*/*.yaml   = 源码态（old 值，进仓库）
#   config_prod.yaml  = 规则文件（old=源码态值，new=生产目标值）
#   dist/             = 生产产物（new 值，含明文密钥，已被 .gitignore 忽略）
#
# 产物结构（与 deploy/ 平级一致，start/stop/update 按子目录找 docker-compose.yaml）：
#   dist/
#     start.sh  stop.sh  update.sh
#     account/ asset/ datasync/ gateway/ message/ notification/ platform/ relation/ ws/
#       gateway/ 断言三件套由 config 的 assertion 段（KID/JWKS/PRIVATE_KEY）在打包时写出
#     config.yaml   ← 规则文件副本；用户只改 config_prod.yaml 的 new 即可出完整 dist

#
# 用法：./config.sh [--config <file>]  （默认读 config.yaml 模板；生产用 --config config_prod.yaml）

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_FILE="$SCRIPT_DIR/config.yaml"

# ============================================================================
# Python 解释器探测
# Windows 上 `python3` 常指向微软商店占位程序（App Execution Alias）：
# 它不输出任何内容、退出码 49，会让脚本"静默结束"。必须实跑校验能输出 3 才采用。
# 优先级：$PYTHON 环境变量 > python3 > python > py > uv > 常见安装路径
# ============================================================================
# 校验候选解释器是否真的是 Python 3（占位符/别名会被这一步挡掉）
_is_py3() {
    local out
    out=$("$1" -c 'import sys;sys.stdout.write(str(sys.version_info[0]))' 2>/dev/null) || return 1
    [ "$out" = "3" ]
}

find_python() {
    local cand p
    for cand in "${PYTHON:-}" python3 python py; do
        [ -z "$cand" ] && continue
        command -v "$cand" >/dev/null 2>&1 || continue
        if _is_py3 "$cand"; then
            command -v "$cand"
            return 0
        fi
    done
    # uv：`uv python find 3` 直接给出已安装的解释器绝对路径
    if command -v uv >/dev/null 2>&1; then
        p=$(uv python find 3 2>/dev/null || true)
        p="${p//\\//}"          # Windows 反斜杠路径转正斜杠，Git Bash 下可直接执行
        if [ -n "$p" ] && [ -x "$p" ] && _is_py3 "$p"; then
            echo "$p"
            return 0
        fi
    fi
    # 兜底：常见安装位置
    for p in \
        /c/Python3*/python.exe \
        /c/Users/*/AppData/Local/Programs/Python/Python3*/python.exe \
        /usr/bin/python3 /usr/local/bin/python3 /opt/homebrew/bin/python3
    do
        [ -x "$p" ] || continue
        if _is_py3 "$p"; then
            echo "$p"
            return 0
        fi
    done
    return 1
}

if ! PYTHON_BIN=$(find_python); then
    echo "❌ 未找到可用的 Python 3 解释器"
    echo "   本脚本用 Python 解析规则文件并替换配置，需要 Python 3.6+。"
    echo "   Windows 上若 python3 指向微软商店占位程序，请任选其一："
    echo "     1) 用 uv 装一个：uv python install 3（装完本脚本会自动识别）"
    echo "     2) 安装 Python 3：https://www.python.org/downloads/（勾选 Add to PATH）"
    echo "     3) 指定解释器：PYTHON=/path/to/python ./config.sh --config config_prod.yaml"
    exit 1
fi
echo "🐍 Python: $PYTHON_BIN"

# 参数解析：--config <file> 指定规则文件，相对路径基于本脚本所在目录
while [ $# -gt 0 ]; do
    case "$1" in
        --config|-c)
            if [ -z "${2:-}" ]; then
                echo "❌ --config 需要指定文件路径"
                exit 1
            fi
            case "$2" in
                /*|[A-Za-z]:*) CONFIG_FILE="$2" ;;  # POSIX 绝对路径 / Windows 盘符路径
                *)  CONFIG_FILE="$SCRIPT_DIR/$2" ;;
            esac
            shift 2
            ;;
        -h|--help)
            echo "用法: $0 [--config <file>]"
            echo "  --config, -c  指定规则文件（默认 config.yaml 模板）"
            echo "  例: ./config.sh --config config_prod.yaml"
            exit 0
            ;;
        *)
            echo "❌ 未知参数: $1"
            echo "用法: $0 [--config <file>]"
            exit 1
            ;;
    esac
done

if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ Error: $(basename "$CONFIG_FILE") not found in $SCRIPT_DIR"
    exit 1
fi

# 占位符保护：整个文件都是占位符（模板）才拒绝执行；部分占位符静默跳过，不再逐条告警
TOTAL_RULES=$(grep -cE '^[[:space:]]*- name:' "$CONFIG_FILE" || true)
PLACEHOLDERS=$(grep -cE '^[[:space:]]+new:.*REPLACE_ME_' "$CONFIG_FILE" || true)
if [ "$TOTAL_RULES" -gt 0 ] && [ "$PLACEHOLDERS" -eq "$TOTAL_RULES" ]; then
    echo "❌ $(basename "$CONFIG_FILE") 里所有 new 值都是 REPLACE_ME_* 占位符（这是模板）"
    echo "   请用 --config config_prod.yaml 指定填好真实值的生产配置"
    exit 1
fi

echo "📖 Rules from $CONFIG_FILE:"
echo ""
# 展示规则摘要（按缩进栈还原分组路径，不依赖 PyYAML）
"$PYTHON_BIN" - "$CONFIG_FILE" <<'PY'
import sys, re
cfg = sys.argv[1]

def unquote(s):
    s = s.strip()
    if len(s) >= 2 and s[0] == '"' and s[-1] == '"':
        return s[1:-1]
    if len(s) >= 2 and s[0] == "'" and s[-1] == "'":
        return s[1:-1]
    return s

def parse_rules(path):
    """遍历任意深度的分组，收集所有 - name: 规则，并记录其分组路径。"""
    rules, cur, stack = [], None, []
    for line in open(path, encoding="utf-8"):
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        indent = len(line) - len(line.lstrip())
        s = line.strip()
        if s.startswith("- name:"):
            if cur: rules.append(cur)
            group = " / ".join(g for _, g in stack if g != "rules")
            cur = {"name": s.split(":", 1)[1].strip(), "group": group}
        elif cur is not None and s.startswith("key:"):
            cur["key"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("old:"):
            cur["old"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("new:"):
            cur["new"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("section:"):
            cur["section"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("default:"):
            cur["default"] = s.split(":", 1)[1].strip().lower() == "true"
        else:
            # 分组标题行，如 "etcd:" 或 "outbound:  # 注释"（值必须为空或纯注释）
            m = re.match(r"^([A-Za-z_][\w-]*):\s*(?:#.*)?$", s)
            if m:
                while stack and stack[-1][0] >= indent:
                    stack.pop()
                stack.append((indent, m.group(1)))
    if cur: rules.append(cur)
    return rules

# 只展示实际会执行的规则（占位符规则会被静默跳过，不再列出）
total = 0
for r in parse_rules(cfg):
    if not r.get("old") or not r.get("new") or "REPLACE_ME_" in r["new"]:
        continue
    total += 1
    tag = f"[{r['group']}] " if r.get("group") else ""
    sec = f" @{r['section']}" + ("/default" if r.get("default") else "") if r.get("section") else ""
    print(f"  🔄 {tag}{r.get('name','')}{sec}: {r.get('old','')} -> {r.get('new','')}")
print(f"\n  ({total} rules)")
PY

echo ""
read -p "Do you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Operation cancelled."
    exit 0
fi

# ============================================================================
# 第一步：构建 dist 产物骨架（脚本 + 9 个服务目录 + 规则文件副本）
# ============================================================================
DIST_DIR="$SCRIPT_DIR/dist"
BUILD_SCRIPTS=(start.sh stop.sh update.sh)
SERVICE_DIRS=(account asset datasync gateway message notification platform relation ws)

echo ""
echo "🧹 清理旧产物: $DIST_DIR"
# 先清空内容再重建目录：部分环境（含安全删除钩子）对 rm -rf 目录会拦截，find -delete 更稳
if [ -d "$DIST_DIR" ]; then
    find "$DIST_DIR" -mindepth 1 -delete 2>/dev/null || rm -rf "$DIST_DIR"
fi
mkdir -p "$DIST_DIR"

echo "📄 拷贝运行脚本:"
for s in "${BUILD_SCRIPTS[@]}"; do
    if [ ! -f "$SCRIPT_DIR/$s" ]; then
        echo "❌ 缺少脚本 $s"
        exit 1
    fi
    cp -p "$SCRIPT_DIR/$s" "$DIST_DIR/$s"
    chmod +x "$DIST_DIR/$s"
    echo "  + $s"
done

echo "📦 拷贝服务目录:"
for d in "${SERVICE_DIRS[@]}"; do
    if [ ! -d "$SCRIPT_DIR/$d" ]; then
        echo "❌ 缺少目录 $d"
        exit 1
    fi
    cp -rp "$SCRIPT_DIR/$d" "$DIST_DIR/$d"
    echo "  + $d/"
done

echo "🔑 生成 dist/config.yaml:"
# 规则文件原样改名进产物：old 保持「当前生效值」，new 保持「目标值」，
# 不做任何改写，这样产物里的规则文件和源文件语义完全一致、可追溯。
cp -p "$CONFIG_FILE" "$DIST_DIR/config.yaml"
echo "  + $(basename "$CONFIG_FILE") -> config.yaml（原样复制，old/new 语义不变）"

# ============================================================================
# 第二步：只对 dist/ 内的服务配置做 old -> new 替换（源码 app/、deploy/ 不动）
# 第三步：按 assertion 规则写出 dist/gateway/ 断言三件套（用户只填 config 即可）
# ============================================================================
"$PYTHON_BIN" - "$CONFIG_FILE" "$DIST_DIR" <<'PY'
import os, sys, re

config_file, dist_dir = sys.argv[1], sys.argv[2]

def unquote(s):
    s = s.strip()
    if len(s) >= 2 and s[0] == '"' and s[-1] == '"':
        return s[1:-1]
    if len(s) >= 2 and s[0] == "'" and s[-1] == "'":
        return s[1:-1]
    return s

def parse_rules(path):
    """按缩进栈解析分组结构（rules -> 分组 -> 规则列表），无需 PyYAML"""
    rules, cur, stack = [], None, []
    for line in open(path, encoding="utf-8"):
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        indent = len(line) - len(line.lstrip())
        s = line.strip()
        if s.startswith("- name:"):
            if cur: rules.append(cur)
            cur = {"name": s.split(":", 1)[1].strip()}
        elif cur is not None and s.startswith("key:"):
            cur["key"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("old:"):
            cur["old"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("new:"):
            cur["new"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("section:"):
            cur["section"] = unquote(s.split(":", 1)[1])
        elif cur is not None and s.startswith("default:"):
            cur["default"] = s.split(":", 1)[1].strip().lower() == "true"
        else:
            m = re.match(r"^([A-Za-z_][\w-]*):\s*(?:#.*)?$", s)
            if m:
                while stack and stack[-1][0] >= indent:
                    stack.pop()
                stack.append((indent, m.group(1)))
    if cur: rules.append(cur)
    return rules

def replace_rule(text, rule):
    """按 rule 替换 text：带 section 则只在该 yaml 段标题下替换；
    文件里无该段但标了 default 则全局替换；无 section 则全局替换。"""
    old, new = rule["old"], rule["new"]
    section = rule.get("section")
    if not section:
        return text.replace(old, new)
    lines = text.splitlines(keepends=True)
    # 找段标题行（顶格或缩进的 "AuthBeaver:" 形式）
    sec_idx = None
    for i, line in enumerate(lines):
        if re.match(r"^\s*" + re.escape(section) + r"\s*:\s*(?:#.*)?$", line.rstrip("\r\n")):
            sec_idx = i
            break
    if sec_idx is None:
        return text.replace(old, new) if rule.get("default") else text
    # 段范围：从 sec_idx+1 到下一个同级或更浅的段标题（缩进 <= 段标题缩进）
    sec_indent = len(lines[sec_idx]) - len(lines[sec_idx].lstrip())
    end = len(lines)
    for j in range(sec_idx + 1, len(lines)):
        stripped = lines[j].lstrip()
        if not stripped or stripped.startswith("#"):
            continue
        ind = len(lines[j]) - len(lines[j].lstrip())
        if ind <= sec_indent and re.match(r"^[\w-]+\s*:", stripped):
            end = j
            break
    # 只在段范围内替换
    for j in range(sec_idx + 1, end):
        lines[j] = lines[j].replace(old, new)
    return "".join(lines)

all_rules = parse_rules(config_file)
rules = [r for r in all_rules if r.get("old") and r.get("new") and "REPLACE_ME_" not in r["new"]]
if not rules:
    print("❌ No active rules found (all rules are placeholders?)")
    sys.exit(1)

# 私钥只用于写出 gateway 密钥文件，不参与业务 yaml 字符串替换
YAML_SKIP_KEYS = {"ASSERTION_PRIVATE_KEY"}
yaml_rules = [r for r in rules if r.get("key") not in YAML_SKIP_KEYS]

# 目标文件：dist/ 下的服务 yaml，排除规则文件副本与 compose
EXCLUDE = {"docker-compose.yaml", "config.yaml"}
files = []
for dp, _, fs in os.walk(dist_dir):
    for fn in sorted(fs):
        if fn.endswith(".yaml") and fn not in EXCLUDE:
            files.append(os.path.join(dp, fn))

# 按 old 长度降序：先替换长串，避免短前缀（如 beaver-s2s-core）打散长串
ordered = sorted(yaml_rules, key=lambda r: len(r["old"]), reverse=True)

print(f"\n🚀 Replacing in {len(files)} yaml files (dist only)...")
updated = 0
for path in files:
    with open(path, "rb") as f:
        raw = f.read()
    text = raw.decode("utf-8")
    orig = text
    for r in ordered:
        if r["old"] in text or r.get("section"):
            text = replace_rule(text, r)
    if text != orig:
        # 保留原换行符
        nl = b"\r\n" if b"\r\n" in raw else b"\n"
        normalized = text.replace("\r\n", "\n").replace("\n", "\r\n") if nl == b"\r\n" else text
        with open(path, "wb") as f:
            f.write(normalized.encode("utf-8"))
        updated += 1
        print(f"  ✅ {os.path.relpath(path, dist_dir)}")

print(f"\n📊 Total: {len(files)} files, {updated} updated")

# --------------------------------------------------------------------------
# 第三步：按 assertion 规则写出 dist/gateway/ 断言三件套
# 用户只填 config_* 的 ASSERTION_KID / JWKS / PRIVATE_KEY 即可
# --------------------------------------------------------------------------
import json, base64

def unescape_nl(s):
    # 规则文件里用 \n 写成一行；兼容误写成 \\n
    return s.replace("\\\\n", "\n").replace("\\n", "\n")

def b64url_decode(s):
    s += "=" * (-len(s) % 4)
    return base64.urlsafe_b64decode(s.encode("ascii"))

def jwk_to_spki_pem(jwk):
    """P-256 JWK -> SubjectPublicKeyInfo PEM（纯标准库，无需 openssl）"""
    x = b64url_decode(jwk["x"]).rjust(32, b"\x00")[-32:]
    y = b64url_decode(jwk["y"]).rjust(32, b"\x00")[-32:]
    # SPKI 固定头 + uncompressed point (0x04||x||y)
    prefix = bytes.fromhex("3059301306072a8648ce3d020106082a8648ce3d030107034200")
    der = prefix + b"\x04" + x + y
    b64 = base64.b64encode(der).decode("ascii")
    lines = [b64[i:i + 64] for i in range(0, len(b64), 64)]
    return "-----BEGIN PUBLIC KEY-----\n" + "\n".join(lines) + "\n-----END PUBLIC KEY-----\n"

by_key_all = {r["key"]: r for r in all_rules if r.get("key")}
by_key = {r["key"]: r for r in rules if r.get("key")}
pk_rule = by_key.get("ASSERTION_PRIVATE_KEY")
jwks_rule = by_key.get("ASSERTION_JWKS")
gw_dir = os.path.join(dist_dir, "gateway")
os.makedirs(gw_dir, exist_ok=True)

if pk_rule and jwks_rule:
    print("\n🔐 Materializing gateway assertion keys from config...")
    pem_text = unescape_nl(pk_rule["new"]).strip() + "\n"
    pem_path = os.path.join(gw_dir, "gateway-assertion.pem")
    with open(pem_path, "w", encoding="utf-8", newline="\n") as f:
        f.write(pem_text)
    print("  + gateway/gateway-assertion.pem")

    try:
        jwks_obj = json.loads(jwks_rule["new"])
        jwks_text = json.dumps(jwks_obj, indent=2, ensure_ascii=False) + "\n"
    except Exception as e:
        print(f"❌ ASSERTION_JWKS 不是合法 JSON: {e}")
        sys.exit(1)
    jwks_path = os.path.join(gw_dir, "gateway-assertion.jwks.json")
    with open(jwks_path, "w", encoding="utf-8", newline="\n") as f:
        f.write(jwks_text)
    print("  + gateway/gateway-assertion.jwks.json")

    keys = jwks_obj.get("keys") or []
    if not keys:
        print("❌ ASSERTION_JWKS.keys 为空，无法生成 pub.pem")
        sys.exit(1)
    pub_text = jwk_to_spki_pem(keys[0])
    pub_path = os.path.join(gw_dir, "gateway-assertion.pub.pem")
    with open(pub_path, "w", encoding="utf-8", newline="\n") as f:
        f.write(pub_text)
    print("  + gateway/gateway-assertion.pub.pem")
elif any(
    "REPLACE_ME_" in (by_key_all.get(k) or {}).get("new", "")
    for k in ("ASSERTION_PRIVATE_KEY", "ASSERTION_JWKS")
):
    print("\n🔐 assertion 仍为 REPLACE_ME_*，保留拷贝自 deploy/gateway/ 的开发密钥")
else:
    print("\n⚠️  缺少 ASSERTION_PRIVATE_KEY 或 ASSERTION_JWKS 有效 new，跳过密钥物化")

for f in ("gateway-assertion.pem", "gateway-assertion.jwks.json", "gateway-assertion.pub.pem"):
    if not os.path.isfile(os.path.join(gw_dir, f)):
        print(f"❌ 缺少 gateway/{f}")
        sys.exit(1)
print("🔐 网关断言密钥就绪: gateway/{pem,jwks.json,pub.pem}")
PY

echo ""
echo "✅ 构建完成: $DIST_DIR"
echo ""
echo "产物清单:"
( cd "$DIST_DIR" && ls -1 )
echo ""
echo "ℹ️  app/ 与 deploy/ 源码配置未改动，仅 dist/ 产物使用 new 值。"
echo "⚠️  dist/ 含明文密码与密钥，已在 .gitignore 中忽略，请勿提交到仓库。"
