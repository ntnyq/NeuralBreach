# 神经裂隙 · Neural Breach

Go + MyGo 原生桌面破解益智游戏。观察十六进制矩阵，从第一行起手，交替沿列、行选码，用有限缓冲区上传重叠的魔偶序列。

- 完整主菜单、四档难度、标准挑战、指定种子自由练习和交互教程。
- 自动保存、暂停遮挡、失焦暂停、退出确认与单槽续玩。
- 分项计分、个人最佳、历史筛选、逐步/自动复盘、解法对照和 JSON / CSV 导出。
- 原创合成电子音乐、事件音效、短促彩色故障碎片，以及静音、减少动态效果、文字缩放设置。
- 简体中文、繁体中文、英语、日语、韩语。内嵌 Rajdhani；中文使用系统字体回退。

## 运行

需要 `go.mod` 指定的 Go 1.27.4 和 MyGo v0.10.0。桌面程序不依赖 Node.js。Linux 运行需要 GTK 3。

```sh
go mod download
make run
# 可选：使用独立测试数据文件
go run . --data ./build/test-data.json
```

第一次启动默认中文。打开「开始游戏 → 新手教程」学习首行限制、行列交替、连续匹配和序列重叠。标准挑战没有提示与撤销，练习成绩不进入个人最佳。

```sh
make dev        # 开发热重建
make build      # 原生应用打包到 build/
make check      # 单元测试、Go vet、翻译检查
make race       # 竞态检查
go tool mygo vet
make snapshots  # 五语言、多尺寸 UI 验证与英文/中文 PNG
```

存档默认位于 `os.UserConfigDir()/neural-breach/data.json`。设置、活动局和历史原子写入同一个文件。损坏文件不会被自动覆盖，可在游戏中备份后重建。`--data` 仅指定游戏数据文件，不改变 MyGo 的进程单实例范围。

## 文档

网站：[neural-breach.ntnyq.dev](https://neural-breach.ntnyq.dev)。

- [玩家与开发入门](docs/getting-started.md)
- [规则与完整设计](docs/game-design.md)
- [架构、存储和验证](docs/architecture.md)
- [打包与发布](docs/releasing.md)
- [设计图与视听预览](docs/design/neural-breach/README.md)

文档站需要 Node.js 和 pnpm 12.21.0：`pnpm install`，然后 `pnpm docs:dev` 或 `pnpm docs:build`。

## 素材

音乐和音效由项目代码原创合成；图标由原创设计稿重建为 SVG，并通过 `make icon` 统一生成应用 PNG、网站图标和 favicon（见[图标生成说明](docs/design/neural-breach/icon-vector.md)）。主菜单背景由内置 imagegen 根据设计稿生成（来源与提示词见 `resources/README.md`）；Rajdhani 使用 SIL OFL，许可位于 `resources/fonts/OFL.txt`。未分发《赛博朋克 2077》的音乐、Logo 或商业中文字体。设计参考图与可播放概念预览保留在 `docs/`，实际游戏以 Go 实现为准。
