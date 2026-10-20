#!/bin/sh

# Only emit whitelisted locale tags to avoid config.js injection from env values.
RUNTIME_DEFAULT_LOCALE=""
case "${DEFAULT_LOCALE:-}" in
  zh-CN|zh-TW|en-US|ru-RU|ko-KR|ja-JP) RUNTIME_DEFAULT_LOCALE="${DEFAULT_LOCALE}" ;;
esac

# 生成运行时配置文件，注入环境变量到前端
FILE_MB=${MAX_FILE_SIZE_MB:-50}
SKILL_MB=${MAX_SKILL_BUNDLE_SIZE_MB:-256}
if [ "$SKILL_MB" -lt "$FILE_MB" ] 2>/dev/null; then
  SKILL_MB=$FILE_MB
fi
if [ "$SKILL_MB" -gt 512 ] 2>/dev/null; then
  SKILL_MB=512
fi

# 保留镜像里 public/config.js 的 API_KEY；WEKNORA_API_KEY 可在部署时覆盖。
RUNTIME_API_KEY="${WEKNORA_API_KEY:-}"
if [ -z "$RUNTIME_API_KEY" ] && [ -f /usr/share/nginx/html/config.js ]; then
  RUNTIME_API_KEY=$(sed -n "s/.*API_KEY: ['\"]\\([^'\"]*\\)['\"].*/\\1/p" /usr/share/nginx/html/config.js | head -1)
fi
RUNTIME_API_KEY=$(printf '%s' "$RUNTIME_API_KEY" | sed 's/\\/\\\\/g; s/"/\\"/g')

cat > /usr/share/nginx/html/config.js << EOF
window.__RUNTIME_CONFIG__ = {
  MAX_FILE_SIZE_MB: ${FILE_MB},
  MAX_SKILL_BUNDLE_SIZE_MB: ${SKILL_MB},
  DEFAULT_LOCALE: "${RUNTIME_DEFAULT_LOCALE}",
  API_KEY: "${RUNTIME_API_KEY}"
};
EOF

# 处理 nginx 配置。
# 两个上限分开注入：全站保持知识库的 MAX_FILE_SIZE，只有技能 zip 上传的两条
# 集合路由放宽到 MAX_SKILL_BUNDLE_SIZE（不含 /install、PATCH 等子路径）。
# 合成一个全站上限会让每个上传端点都能收到技能包那么大的 body。
export MAX_FILE_SIZE=${FILE_MB}M
export MAX_SKILL_BUNDLE_SIZE=${SKILL_MB}M
export APP_HOST=${APP_HOST:-app}
export APP_PORT=${APP_PORT:-8080}
export APP_SCHEME=${APP_SCHEME:-http}
envsubst '${MAX_FILE_SIZE} ${MAX_SKILL_BUNDLE_SIZE} ${APP_HOST} ${APP_PORT} ${APP_SCHEME}' \
  < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf

# 启动 nginx
exec nginx -g 'daemon off;'
