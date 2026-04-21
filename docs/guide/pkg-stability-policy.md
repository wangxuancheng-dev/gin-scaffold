# pkg 稳定性治理策略

本页定义 `pkg/*` 的版本演进约束，配合 `pkg/STABILITY.yaml` 与 CI 守卫执行。

## 稳定性等级

- `stable`
  - 目标：对外可依赖，默认承诺向后兼容。
  - 规则：公开 API（导出函数、导出结构体字段、错误语义）变更必须提供兼容期。
- `experimental`
  - 目标：快速迭代验证。
  - 规则：允许破坏性变更，但必须在 PR 与发布说明标注 breaking change。

## 兼容窗口与弃用流程

- 兼容窗口（`stable`）：
  - 至少一个小版本周期保持旧 API 可用（可标记 deprecated）。
  - 删除旧 API 前，需完成代码内替换与文档迁移说明。
- 弃用流程（`stable`）：
  1. 在注释里标记 `Deprecated: use <new API>`.
  2. 在变更记录中写明迁移路径。
  3. 在下一个版本窗口后再删除。

## 变更审批要求

- `stable` 包变更需附：
  - 影响面（调用方/行为变化）
  - 回滚策略
  - 迁移步骤
- 推荐使用模板：`docs/guide/pkg-stability-change-template.md`。

## 与 CI 的关系

- `scripts/check-pkg-stability.sh` 负责目录与稳定性清单一致性。
- 本策略负责“如何改”，脚本负责“是否声明”。
