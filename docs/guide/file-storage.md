# 文件存储（V1）

当前版本提供基础文件存储能力（本地磁盘或 S3 兼容对象存储）：

- 上传：`POST /api/v1/client/files/upload`（JWT）
- 直传预签名（仅 `s3`/`minio`）：`POST /api/v1/client/files/presign`（JWT）
- 生成签名下载地址：`GET /api/v1/client/files/*key/url`（JWT）
- 下载：`GET /api/v1/client/files/*key/download?e=<unix>&sig=<hmac>`

## 上传约定

- 请求类型：`multipart/form-data`
- 文件字段：`file`
- 扩展名白名单由 `storage.allowed_ext` 控制
- 内容类型白名单由 `storage.allowed_mime` 控制（读取文件头后使用 `http.DetectContentType` 嗅探）
- 大小上限由 `storage.max_upload_mb` 控制

## 接口联调示例

假设服务地址为 `http://localhost:8080`，并且已拿到 JWT：

```bash
TOKEN="<your_access_token>"
```

1) 上传文件：

```bash
curl -X POST "http://localhost:8080/api/v1/client/files/upload" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "file=@./demo.txt"
```

返回中会包含 `data.key`。

2) 获取签名下载地址：

```bash
curl "http://localhost:8080/api/v1/client/files/<key>/url?expire_sec=300" \
  -H "Authorization: Bearer ${TOKEN}"
```

返回中会包含 `data.url`（相对路径）。

3) 下载文件：

```bash
curl -L "http://localhost:8080<signed_url>" -o downloaded-demo.txt
```

### S3/MinIO 直传（预签名 PUT）

1) 申请预签名：

```bash
curl -X POST "http://localhost:8080/api/v1/client/files/presign?expire_sec=600" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"filename\":\"demo.txt\",\"content_type\":\"text/plain\"}"
```

2) 使用返回的 `data.method` / `data.url` / `data.headers` 直传对象：

```bash
curl -X PUT "<presigned_url>" \
  -H "Content-Type: text/plain" \
  --data-binary @./demo.txt
```

## Swagger

- 文档页：`/swagger/index.html`
- 文件接口分组：`client-file`

## 配置示例

```yaml
storage:
  enabled: true
  driver: local
  local_dir: ./storage
  sign_secret: "replace-me"
  max_upload_mb: 10
  allowed_ext: ".jpg,.jpeg,.png,.pdf,.txt"
  allowed_mime: "image/jpeg,image/png,application/pdf,text/plain"
  url_expire_sec: 300
```

### MinIO / S3 示例

```yaml
storage:
  enabled: true
  driver: minio
  sign_secret: "replace-me"
  max_upload_mb: 20
  allowed_ext: ".jpg,.jpeg,.png,.pdf,.txt"
  allowed_mime: "image/jpeg,image/png,application/pdf,text/plain"
  url_expire_sec: 300
  s3_endpoint: "https://minio.example.com"
  s3_region: "us-east-1"
  s3_bucket: "gin-scaffold"
  s3_access_key: "${S3_ACCESS_KEY}"
  s3_secret_key: "${S3_SECRET_KEY}"
  s3_path_style: true
  s3_insecure: false
```

## 安全建议

- 生产环境务必替换 `storage.sign_secret`
- 上传目录放在应用工作目录外或独立挂载盘
- 下载链接默认短时有效，按业务按需调小
