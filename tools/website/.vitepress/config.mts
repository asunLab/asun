import { defineConfig } from 'vitepress'
import { readFileSync } from 'fs'
import { resolve } from 'path'

// 本地开发: /
// GitHub 项目页: /ason/  (由 CI 通过 VITE_BASE 环境变量注入)
// 自定义域名: /
const BASE = process.env.VITE_BASE ?? '/'

const asonGrammar = JSON.parse(
  readFileSync(
    resolve(__dirname, '../../vscode_ason_plugins/syntaxes/ason.tmLanguage.json'),
    'utf-8'
  )
)

const enNav = [
  { text: 'Guide', link: '/guide/what-is-ason' },
  { text: 'Syntax', link: '/reference/syntax' },
  { text: 'Performance', link: '/reference/performance' },
  {
    text: 'Languages',
    items: [
      { text: 'Rust', link: '/languages/rust' },
      { text: 'Go', link: '/languages/go' },
      { text: 'Python', link: '/languages/python' },
      { text: 'C', link: '/languages/c' },
      { text: 'C++', link: '/languages/cpp' },
      { text: 'Java', link: '/languages/java' },
      { text: 'Zig', link: '/languages/zig' },
    ],
  },
  { text: 'Spec', link: '/spec' },
]

const zhNav = [
  { text: '指南', link: '/zh/guide/what-is-ason' },
  { text: '语法', link: '/zh/reference/syntax' },
  { text: '性能', link: '/zh/reference/performance' },
  {
    text: '语言',
    items: [
      { text: 'Rust', link: '/zh/languages/rust' },
      { text: 'Go', link: '/zh/languages/go' },
      { text: 'Python', link: '/zh/languages/python' },
      { text: 'C', link: '/zh/languages/c' },
      { text: 'C++', link: '/zh/languages/cpp' },
      { text: 'Java', link: '/zh/languages/java' },
      { text: 'Zig', link: '/zh/languages/zig' },
    ],
  },
  { text: '规范', link: '/zh/spec' },
]

const enSidebar = {
  '/guide/': [
    {
      text: 'Introduction',
      items: [
        { text: 'What is ASON?', link: '/guide/what-is-ason' },
        { text: 'Getting Started', link: '/guide/getting-started' },
        { text: 'Why ASON over JSON?', link: '/guide/why-ason' },
      ],
    },
    {
      text: 'Concepts',
      items: [
        { text: 'Schema & Data', link: '/guide/schema-and-data' },
        { text: 'LLM Integration', link: '/guide/llm-integration' },
        { text: 'Binary Format', link: '/guide/binary-format' },
      ],
    },
  ],
  '/reference/': [
    {
      text: 'Reference',
      items: [
        { text: 'Syntax', link: '/reference/syntax' },
        { text: 'Data Types', link: '/reference/data-types' },
        { text: 'Performance', link: '/reference/performance' },
      ],
    },
  ],
  '/languages/': [
    {
      text: 'Language Guides',
      items: [
        { text: 'Rust', link: '/languages/rust' },
        { text: 'Go', link: '/languages/go' },
        { text: 'Python', link: '/languages/python' },
        { text: 'C', link: '/languages/c' },
        { text: 'C++', link: '/languages/cpp' },
        { text: 'Java', link: '/languages/java' },
        { text: 'Zig', link: '/languages/zig' },
      ],
    },
  ],
}

const zhSidebar = {
  '/zh/guide/': [
    {
      text: '入门',
      items: [
        { text: 'ASON 是什么？', link: '/zh/guide/what-is-ason' },
        { text: '快速开始', link: '/zh/guide/getting-started' },
        { text: '为什么选择 ASON？', link: '/zh/guide/why-ason' },
      ],
    },
    {
      text: '核心概念',
      items: [
        { text: 'Schema 与数据', link: '/zh/guide/schema-and-data' },
        { text: 'LLM 集成', link: '/zh/guide/llm-integration' },
        { text: '二进制格式', link: '/zh/guide/binary-format' },
      ],
    },
  ],
  '/zh/reference/': [
    {
      text: '参考文档',
      items: [
        { text: '语法参考', link: '/zh/reference/syntax' },
        { text: '数据类型', link: '/zh/reference/data-types' },
        { text: '性能基准', link: '/zh/reference/performance' },
      ],
    },
  ],
  '/zh/languages/': [
    {
      text: '语言指南',
      items: [
        { text: 'Rust', link: '/zh/languages/rust' },
        { text: 'Go', link: '/zh/languages/go' },
        { text: 'Python', link: '/zh/languages/python' },
        { text: 'C', link: '/zh/languages/c' },
        { text: 'C++', link: '/zh/languages/cpp' },
        { text: 'Java', link: '/zh/languages/java' },
        { text: 'Zig', link: '/zh/languages/zig' },
      ],
    },
  ],
}

export default defineConfig({
  base: BASE,
  title: 'ASON',
  titleTemplate: ':title — ASON',

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#646cff' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'ASON' }],
    ['meta', { property: 'og:image', content: '/og-image.png' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: '/og-image.png' }],
  ],

  sitemap: {
    hostname: 'https://ason-format.dev',
  },

  markdown: {
    shikiSetup: async (shiki) => {
      await shiki.loadLanguage({
        ...asonGrammar,
        name: 'ason',
        scopeName: 'source.ason',
      } as any)
    },
  },

  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      description:
        'ASON (Array-Schema Object Notation) — a high-performance serialization format with 65% token savings over JSON, ideal for LLM interaction and large-scale data exchange.',
      themeConfig: {
        nav: enNav,
        sidebar: enSidebar,
        editLink: {
          pattern: 'https://github.com/your-org/ason/edit/main/tools/website/:path',
          text: 'Edit this page on GitHub',
        },
        footer: {
          message: 'Released under the MIT License.',
          copyright: 'Copyright © 2024-present ASON Contributors',
        },
      },
    },
    zh: {
      label: '中文',
      lang: 'zh-CN',
      link: '/zh/',
      description:
        'ASON（数组模式对象符号）—— 高性能序列化格式，相比 JSON 节省 65% Token，专为 LLM 交互与大规模数据传输设计。',
      themeConfig: {
        nav: zhNav,
        sidebar: zhSidebar,
        editLink: {
          pattern: 'https://github.com/your-org/ason/edit/main/tools/website/:path',
          text: '在 GitHub 上编辑此页',
        },
        footer: {
          message: '基于 MIT 许可证发布。',
          copyright: 'Copyright © 2024-present ASON Contributors',
        },
        docFooter: {
          prev: '上一页',
          next: '下一页',
        },
        outline: {
          label: '本页目录',
        },
        lastUpdated: {
          text: '最后更新于',
        },
        returnToTopLabel: '回到顶部',
        sidebarMenuLabel: '菜单',
        darkModeSwitchLabel: '主题',
      },
    },
  },

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'ASON',

    socialLinks: [
      { icon: 'github', link: 'https://github.com/your-org/ason' },
    ],

    search: {
      provider: 'local',
      options: {
        locales: {
          zh: {
            translations: {
              button: { buttonText: '搜索', buttonAriaLabel: '搜索' },
              modal: {
                noResultsText: '无法找到相关结果',
                resetButtonTitle: '清除查询条件',
                footer: { selectText: '选择', navigateText: '切换', closeText: '关闭' },
              },
            },
          },
        },
      },
    },
  },
})
