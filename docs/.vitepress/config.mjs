import { defineConfig } from 'vitepress'

export default defineConfig({
  title: '神经裂隙 · Neural Breach',
  description: 'Neural Breach 原生桌面破解游戏',
  lang: 'zh-CN',
  lastUpdated: true,
  sitemap: {
    hostname: 'https://neural-breach.ntnyq.dev',
  },
  transformHead({ siteData }) {
    return [
      [
        'link',
        {
          rel: 'icon',
          href: `${siteData.base}favicon.ico`,
          sizes: '16x16 32x32 48x48 64x64',
        },
      ],
      [
        'link',
        {
          rel: 'icon',
          href: `${siteData.base}icon.svg`,
          type: 'image/svg+xml',
          sizes: 'any',
        },
      ],
    ]
  },
  themeConfig: {
    logo: { src: '/icon.png', alt: '神经裂隙 · Neural Breach' },
    nav: [
      { text: '入门', link: '/getting-started' },
      { text: '规则', link: '/game-design' },
      { text: '架构', link: '/architecture' },
    ],
    sidebar: [
      { text: '开始入侵', link: '/getting-started' },
      { text: '完整规则与设计', link: '/game-design' },
      { text: '架构与数据', link: '/architecture' },
      { text: '构建与发布', link: '/releasing' },
      { text: '设计图与视听预览', link: '/design/neural-breach/' },
      { text: '字体研究', link: '/design/neural-breach/typography' },
    ],
    outline: { label: '本页目录', level: [2, 3] },
    search: { provider: 'local' },
  },
})
