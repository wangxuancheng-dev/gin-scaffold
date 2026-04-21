import { defineConfig } from "vitepress";

export default defineConfig({
  title: "gin-scaffold",
  description: "Gin Scaffold 企业级中小团队生产文档",
  lang: "zh-CN",
  cleanUrls: true,
  markdown: {
    lineNumbers: true,
  },
  themeConfig: {
    nav: [
      { text: "开发手册", link: "/guide/handbook" },
      { text: "指南", link: "/guide/introduction" },
      { text: "阅读路径", link: "/paths/developer" },
      { text: "部署", link: "/ops/production-runbook" },
      { text: "检查清单", link: "/checklist" },
      { text: "GitHub", link: "https://github.com/wangxuancheng-dev/gin-scaffold" },
    ],
    sidebar: {
      "/guide/": [
        {
          text: "开发手册",
          items: [{ text: "手册总览与索引", link: "/guide/handbook" }],
        },
        {
          text: "入门",
          items: [
            { text: "项目简介", link: "/guide/introduction" },
            { text: "新人入门（架构 + FAQ）", link: "/guide/onboarding" },
            { text: "快速开始", link: "/guide/quick-start" },
          ],
        },
        {
          text: "架构与 HTTP",
          items: [
            { text: "目录结构与分层", link: "/guide/directory-structure" },
            { text: "路由与分组", link: "/guide/routing" },
            { text: "中间件参考", link: "/guide/middleware-reference" },
            { text: "错误与响应", link: "/guide/error-handling" },
          ],
        },
        {
          text: "配置",
          items: [
            { text: "配置说明（关键组）", link: "/guide/configuration" },
            { text: "配置详解（全量键）", link: "/guide/configuration-advanced" },
          ],
        },
        {
          text: "数据、缓存、队列",
          items: [
            { text: "数据库迁移与填充", link: "/guide/database-and-migrations" },
            { text: "缓存使用", link: "/guide/caching" },
            { text: "异步队列（Asynq）", link: "/guide/queues-asynq" },
            { text: "定时任务中心", link: "/guide/scheduler" },
          ],
        },
        {
          text: "命令行与代码生成",
          items: [
            { text: "命令系统", link: "/guide/commands" },
            { text: "代码生成（CRUD）", link: "/guide/codegen" },
            { text: "生成器走读", link: "/guide/codegen-walkthrough" },
          ],
        },
        {
          text: "文件、实时、国际化、限流",
          items: [
            { text: "文件存储", link: "/guide/file-storage" },
            { text: "SSE 与 WebSocket", link: "/guide/realtime-sse-websocket" },
            { text: "本地化（i18n）", link: "/guide/i18n" },
            { text: "验证器与多语言", link: "/guide/validation-i18n" },
            { text: "全局限流", link: "/guide/rate-limiting" },
          ],
        },
        {
          text: "横切、可观测、安全、测试",
          items: [
            { text: "平台横切能力", link: "/guide/platform" },
            { text: "出站 HTTP 客户端", link: "/guide/outbound-httpclient" },
            { text: "日志与可观测", link: "/guide/logging-observability" },
            { text: "安全实践", link: "/guide/security-practices" },
            { text: "RBAC 与权限", link: "/guide/rbac-and-permissions" },
            { text: "管理端 API 总览", link: "/guide/admin-api-overview" },
            { text: "测试指南", link: "/guide/testing-guide" },
            { text: "常用包与辅助能力", link: "/guide/helpers-packages" },
          ],
        },
        {
          text: "参考与附录",
          items: [
            { text: "环境变量绑定一览", link: "/guide/environment-variables" },
            { text: "数据库与 GORM 实践", link: "/guide/database-patterns" },
          ],
        },
      ],
      "/ops/": [
        {
          text: "运维与上线",
          items: [
            { text: "生产运行手册", link: "/ops/production-runbook" },
            { text: "上线检查清单", link: "/checklist" },
          ],
        },
      ],
      "/paths/": [
        {
          text: "按角色阅读",
          items: [
            { text: "开发同学", link: "/paths/developer" },
            { text: "运维同学", link: "/paths/operations" },
            { text: "测试同学", link: "/paths/testing" },
          ],
        },
      ],
      "/meta/": [
        {
          text: "文档维护",
          items: [{ text: "文档贡献规范", link: "/meta/docs-maintenance" }],
        },
      ],
    },
    socialLinks: [{ icon: "github", link: "https://github.com/your-org/gin-scaffold" }],
    search: {
      provider: "local",
    },
    outline: {
      label: "本页目录",
      level: [2, 3],
    },
    docFooter: {
      prev: "上一页",
      next: "下一页",
    },
    darkModeSwitchLabel: "外观",
    sidebarMenuLabel: "菜单",
    returnToTopLabel: "回到顶部",
  },
});
