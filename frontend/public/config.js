// 运行时配置（本地开发默认值，Docker 环境会被 entrypoint 脚本覆盖）
window.__RUNTIME_CONFIG__ = {
  MAX_FILE_SIZE_MB: 50,
  MAX_SKILL_BUNDLE_SIZE_MB: 256,
  // Optional: serve embed on a dedicated origin, e.g. 'https://embed.example.com'
  EMBED_BASE_URL: '',
  // Optional: default UI locale for first-time visitors (zh-CN | zh-TW | en-US | ru-RU | ko-KR | ja-JP)
  DEFAULT_LOCALE: '',
  // 非空时，除嵌入页外的请求改走 X-API-Key，不再附带登录 Bearer 和 X-Tenant-ID。留空则走原来的登录态。
  API_KEY: 'sk-a7LlhCRfXV-uOgpXqnpV1uIaXgxpclSa8dyhKDJ-zIc',
};
