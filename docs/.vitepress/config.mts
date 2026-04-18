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
      { text: "指南", link: "/guide/introduction" },
      { text: "阅读路径", link: "/paths/developer" },
      { text: "部署", link: "/ops/production-runbook" },
      { text: "检查清单", link: "/checklist" },
      { text: "GitHub", link: "https://github.com/your-org/gin-scaffold" },
    ],
    sidebar: {
      "/guide/": [
        {
          text: "快速开始",
          items: [
            { text: "项目简介", link: "/guide/introduction" },
            { text: "新人入门（架构 + FAQ）", link: "/guide/onboarding" },
            { text: "快速开始", link: "/guide/quick-start" },
            { text: "配置说明", link: "/guide/configuration" },
            { text: "命令系统", link: "/guide/commands" },
          ],
        },
        {
          text: "功能模块",
          items: [
            { text: "定时任务中心", link: "/guide/scheduler" },
            { text: "日志与可观测", link: "/guide/logging-observability" },
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
