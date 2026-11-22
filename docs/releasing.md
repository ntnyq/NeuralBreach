# 构建与发布

应用名称与标识在 `mygo.json`，版本同时在 `main.go` 的 `version` 常量中维护。

应用名为 `Neural Breach`（神经裂隙），网站为 `https://neural-breach.ntnyq.dev`。文档站构建命令为 `pnpm docs:build`，输出目录为 `docs/.vitepress/dist`；部署时将该域名绑定到静态站点，并配置对应的 DNS 记录。`docs/public/CNAME` 用于 GitHub Pages 自定义域名，站点地图域名在 `docs/.vitepress/config.mjs` 中配置。

```sh
make check
make race
go tool mygo vet
make build
# 只构建 macOS 应用，不生成 DMG
make icon-check
go tool mygo build -skip-dmg
```

输出位于 `build/`。字体通过 Go embed 打包，音乐通过代码生成，没有运行时字体或音频下载。

Windows 发布包只使用原生 UI。手动打包时使用 `GOFLAGS=-tags=mygo_nowebview go tool mygo build -platform windows/amd64`，构建产物会省去 WebView2 加载代码。

应用图标和文档站图标统一来自 `resources/icon-macos.svg`。修改后使用 resvg 0.50.0 执行 `make icon`，并提交所有生成资源及清单；详见[图标生成说明](./design/neural-breach/icon-vector.md)。`make build` 和 CI/发布工作流在打包前自动校验图标是否同步，MyGo 从 `resources/icon.png` 生成各平台图标。正常打包无需安装 resvg；直接调用 MyGo CLI 时，先执行 `make icon-check`。

`.github/workflows/ci.yml` 检查格式、翻译生成一致性、竞态和 MyGo vet，并交叉编译 Windows/Linux、打包 macOS。发布工作流支持手动输入稳定版本号，生成更新日志、三平台安装包和校验和。工作流中的发布步骤会推送版本提交与 tag，并发布 GitHub Release；本地 `make build` 不会发布任何内容。

macOS 本地开发构建使用临时签名。正式分发前需要自行配置 Developer ID、签名和公证；Windows 签名也需要发布者证书。Linux 桌面运行依赖 GTK 3。当前游戏没有应用内自动更新界面。

## 发布前人工检查

- 五种语言下菜单、8×8 棋盘、目标滚动、键盘焦点及 150% 缩放。
- 失焦与最小化自动暂停，恢复后不自动计时。
- 播放、静音、暂停淡出、关闭资源释放及无音频设备降级。
- 系统关闭按钮和菜单退出均等待保存；模拟不可写目录可重试或取消。
- 记录 JSON/CSV 导出、同种子练习、复盘实际路径和验证解法。
- 在每个目标系统实际运行安装包；交叉编译成功仅证明能构建。
