# 验证器与多语言

本项目使用 `go-playground/validator` + `i18n/*.json`

## 核心机制

- 规则写在请求 DTO 的 `binding` 标签里。
- handler 里统一调用 `handler.FailInvalidParam(c, err)` 输出校验错误。
- 错误消息优先从 `i18n` 读取，支持字段级覆盖。
- `Accept-Language` 自动选择中英文（`en*` 走英文，其它默认中文）。

## 规则写法（基础）

```go
type UserCreateRequest struct {
    Username string `json:"username" binding:"required,min=3,max=32,not_admin"`
    Password string `json:"password" binding:"required,min=6,max=64"`
    Role     string `json:"role" binding:"omitempty,oneof=admin user"`
}
```

## 规则写法（跨字段）

> 注意：跨字段参数使用 **Go 字段名**，不是 json 字段名。

```go
type TimeRange struct {
    StartAt time.Time `json:"start_at" binding:"required"`
    EndAt   time.Time `json:"end_at" binding:"required,after_field=StartAt"`
}

type PasswordReset struct {
    Password        string `json:"password" binding:"required,min=6"`
    ConfirmPassword string `json:"confirm_password" binding:"required,same_field=Password"`
}
```

## 常用内置规则（建议）

- `required`：必填
- `omitempty`：为空则跳过后续规则
- `min` / `max`：最小/最大长度（字符串）或最小/最大值（数字）
- `len`：固定长度
- `oneof=a b c`：枚举
- `email`：邮箱格式
- `gte` / `lte`：大于等于 / 小于等于

更多规则可参考 [go-playground/validator 文档](https://pkg.go.dev/github.com/go-playground/validator/v10)。

## 项目内置自定义规则

- `not_admin`：字符串字段不能是保留值 `admin`
- `same_field=<Field>`：当前字段必须与指定字段一致
- `after_field=<Field>`：当前字段必须晚于指定字段（支持时间/可比较值）

自定义规则定义文件：`internal/pkg/validator/validator_rules.go`

## i18n key 解析顺序

以 `username.min` 为例，按顺序命中：

1. `validation.custom.username.min`
2. `validation.custom.*.min`
3. `validation.username.min`（兼容键）
4. `validation.min`
5. validator 默认翻译
6. `validation.invalid` / 内置默认模板（最终兜底）

## 推荐语言包键

`i18n/zh.json` / `i18n/en.json` 至少建议包含：

- `validation_failed`（整体前缀）
- `validation.required` / `validation.min` / `validation.max` / `validation.oneof` / `validation.email` ...
- `validation.attributes.<field>`（字段别名）
- `validation.custom.<field>.<rule>`（字段级覆盖）
- `validation.invalid`（兜底消息）

示例（中文）：

```json
{
  "validation_failed": "参数校验失败: {details}",
  "validation.min": ":attribute 长度不能少于 :min",
  "validation.attributes.username": "用户名",
  "validation.custom.username.min": "用户名长度不能少于 :min",
  "validation.custom.username.not_admin": "用户名不能使用 admin"
}
```

## 占位符

支持两种风格（可混用）：

- 冒号风格：`:attribute`、`:min`、`:max`、`:value`、`:other`
- 花括号风格：`{details}`、`{field}`、`{param}`

## 错误响应结构

校验失败统一返回：

- `msg`：总览消息（可读）
- `data.errors[]`：结构化详情
  - `field`
  - `rule`
  - `message`

典型响应：

```json
{
  "code": 400,
  "msg": "参数校验失败: 用户名长度不能少于 3",
  "data": {
    "errors": [
      {"field":"username","rule":"min","message":"用户名长度不能少于 3"}
    ]
  }
}
```

## 新增自定义规则步骤

1. 在 `internal/pkg/validator/validator_rules.go` 的分组函数中新增 rule spec（字符串/数值/跨字段）。
2. 在 `i18n/zh.json`、`i18n/en.json` 增加：
   - `validation.<rule>`
   - 或字段级 `validation.custom.<field>.<rule>`
3. 给规则补单测（`internal/pkg/validator/validator_test.go`）。
4. 若涉及具体接口，再补 handler 侧响应测试。

## 常见排障

- 命中了默认文案而不是自定义文案：检查 key 是否写成 `validation.custom.<field>.<rule>`。
- 跨字段规则未生效：检查参数是否用了 Go 字段名（如 `StartAt`），不是 `start_at`。
- 英文不生效：确认请求头 `Accept-Language` 以 `en` 开头。
