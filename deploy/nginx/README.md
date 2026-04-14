# Nginx Deployment Notes

本文档配合 `deploy/nginx/gin-scaffold.conf.example` 使用，目标是 5 分钟完成单机反向代理配置。

## 1. 准备前提

- 应用已通过 `systemd` 运行在 `127.0.0.1:8080`
- 域名已解析到服务器公网 IP
- 已安装 Nginx
- 已准备 TLS 证书（推荐 Let's Encrypt）

## 2. 快速落地步骤

1) 复制模板：

```bash
sudo cp deploy/nginx/gin-scaffold.conf.example /etc/nginx/conf.d/gin-scaffold.conf
```

2) 修改以下关键项：

- `server_name example.com` -> 你的域名
- `ssl_certificate` / `ssl_certificate_key` -> 实际证书路径
- `/metrics` 与 `/swagger/` 中的 `allow` 网段 -> 你的内网/运维出口 IP

3) 校验并重载：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 3. 验证清单

- `https://<your-domain>/health` 返回成功
- 业务 API 可访问
- `http://<your-domain>` 自动跳转到 HTTPS
- 非白名单来源访问 `/metrics` 返回拒绝

## 4. 常见调整项

- **上传文件较大**：调大 `client_max_body_size`
- **长请求超时**：调整 `proxy_read_timeout` / `proxy_send_timeout`
- **只内网开放 Swagger**：保留 `/swagger/` 白名单策略

## 5. 生产建议

- 后端应用仅监听内网地址（如 `127.0.0.1`）
- DB/Redis 不暴露公网
- Nginx access/error 日志接入轮转与监控
