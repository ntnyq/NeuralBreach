# 设计图生成提示词

补充：[缓冲格数与字体修订提示词](./revision-prompts.md)。

工具：内置 `image_gen`；全部使用非透明背景。以下为实际发送的完整提示词，局部修订保留在末尾。最终图片保存在 `docs/public/neural-breach/`。

## 01-main-menu

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: real videogame MAIN MENU. Strong asymmetrical composition: left 38% spacious vertical menu; right 62% dark atmospheric industrial cyberdeck and tangled cable silhouettes, with dim scarlet circuitry and cyan signal fragments, low contrast so menu is focus. Top-left large yellow game wordmark at x~100,y~140. Under title small red subtitle “入侵协议训练终端”. At left center a vertical list of FIVE large menu actions, each on its own row: “开始游戏”, “继续游戏”, “查看记录”, “设置”, “退出游戏”. “开始游戏” is currently selected with a cyan horizontal translucent rectangular band, cyan outline and left chevron, other options coral red. Make these look exactly like a serious AAA PC game's start menu. Under “继续游戏” small readable save subtitle “标准 · 缓冲区 3/8 · 剩余 42.00 秒”. Bottom-left small “本地存档已同步”. Bottom-right keyboard hints “↑↓ 选择    Enter 确认”. Far top-right small build “v0.1”. Composition has cinematic negative space, not a grid of widgets.
```

## 03-breach-game

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: active BREACH PROTOCOL gameplay, reference authentic Cyberpunk hacking overlay proportions. Top-left yellow heading “入侵协议”; below a red timer label “剩余时间” and huge yellow “42.00”. Thin timer bar. Top-right small “标准” and “暂停 Esc”.
Upper center/right horizontal “缓冲区 3/8”: exactly EIGHT equal square slots, first THREE containing cyan “1C” “55” “BD”, followed by five empty outline slots.
Left 56%: heading “代码矩阵”, a precisely aligned 6x6 grid of large readable hex codes, no extra cells. Matrix rows:
55 1C E9 BD 7A FF
BD E9 55 1C FF FF
7A 55 FF E9 BD 1C
FF BD 1C 55 E9 7A
E9 FF 7A 1C E9 BD
1C 7A BD FF 55 E9
Mark exactly three already used cells: row1 col2 “1C” with tiny step 1, row3 col2 “55” with tiny step 2, row3 col5 “BD” with tiny step 3; used cells dim and crossed by a small diagonal, but text remains readable. Current legal choices are unused cells of COLUMN 5 only; highlight column5 with translucent yellow stripe and thin vertical lines. Its row3 cell is already used, so darkened. Focus row5 col5 E9 with cyan corner brackets. Under matrix clear yellow instruction “下一步：在第 5 列选择”.
Right 40%: heading “待上传魔偶”, three vertically arranged flat rows:
“数据虹吸” — code “1C 55 BD” — cyan check and “已上传”
“监控静默” — code “BD E9 7A” — first BD cyan and remaining yellow — “1/3”
“痕迹抹除” — code “7A 55 FF” — “等待”
Below small “已上传 1/3” and “基础分 450”. Lower edge hints “方向键 移动    Enter 选码    Esc 暂停”. A small subdued outlined action “结束并结算” at bottom right. No decorative illustration over board; matrix and target codes dominate. Serious crisp dark red/yellow/cyan real game HUD.
```

## 02-new-game

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: NEW GAME / difficulty selection. This is a full screen game submenu, with “开始游戏” as a large yellow title top left, red rule below, small brand top-right. Upper tabs “标准挑战” selected cyan and “自由练习” unselected red. Main region left 58% has FOUR wide stacked angular rows, selectable like Cyberpunk's difficulty menu, not cards. Text EXACTLY:
“入门” / “5×5 矩阵 · 7 格缓冲 · 90 秒”
“标准” / “6×6 矩阵 · 8 格缓冲 · 60 秒”
“困难” / “7×7 矩阵 · 9 格缓冲 · 45 秒”
“极限” / “8×8 矩阵 · 10 格缓冲 · 35 秒”
Select 标准 with pale cyan fill and black text, other rows scarlet type and faint red outlines. Right region separated by long thin red vertical line: huge yellow “标准”, small “适合熟悉行列切换的黑客”, a refined wireframe cyberdeck circuit illustration, then readable feature rows “3 个魔偶”, “序列长度 3–4”, “得分倍率 ×1.5”. Lower explanatory text “从第一行开始，交替选择列与行。首次选码后开始计时。” Large yellow primary angular bottom-right button “接入网络”. Secondary bottom-left “Esc 返回”. Small tertiary bottom-center “新手教程”. Dark black-burgundy background with almost invisible red circuit noise. Spacious console-game settings hierarchy.
```

## 04-results

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: full success RESULTS / debrief game screen. Top-left yellow “入侵完成”, cyan check motif and smaller “全部魔偶上传成功”. Top-right small “标准 · 标准挑战”. Left 40% a monumental yellow “2,790” score with label “本局得分”, cyan compact stamp “个人新纪录”, then “上传 3/3” “用时 24.00 秒” “缓冲区 7/8”. Right 55% a flat framed red-bordered receipt table titled “得分明细”, exactly these rows in ivory with right-aligned yellow values:
“魔偶基础分” “900”
“多目标奖励” “300”
“全部上传” “400”
“剩余时间” “180”
“缓冲节约” “80”
thin red horizontal divider,
“基础合计” “1,860”
“标准倍率” “×1.5”
“最终得分” “2,790”.
Below receipt cyan “记录已保存到本地”. Across bottom a path strip labeled “上传路径” with seven hex tokens “1C 55 BD E9 7A 55 FF”, connected subtly. Bottom actions: yellow primary “再来一局”; red outline “查看复盘”; red text “返回主菜单”. Serious restrained cyberpunk debrief, no confetti, no medals, no illustration crowding numbers.
```

## 05-records

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: local game HISTORY and selected run detail in split view. Large yellow title top-left “入侵记录”. Top navigation thin horizontal tabs “历史记录” selected cyan, “个人最佳” muted red. Under them two compact filters “标准挑战” and “标准”; right small “仅保存在本机”.
Left 64% a dense game archive table with thin red horizontal lines, NO rounded cards. Column headings “时间”, “难度”, “上传”, “分数”, “结果”. Five realistic rows:
“10-23 21:48” “标准” “3/3” “2,790” “全部上传”
“10-23 21:42” “标准” “2/3” “1,125” “部分上传”
“10-23 21:38” “标准” “0/3” “0” “入侵失败”
“10-23 21:31” “标准” “3/3” “2,655” “全部上传”
“10-23 21:26” “标准” “1/3” “450” “部分上传”
first row selected with cyan border and very subtle cyan wash, result labels cyan or red.
Right 32% detail pane separated by red line: “记录详情”, huge yellow “2,790”, “标准 · 全部上传”, “用时 24.00 秒”, “缓冲区 7/8”, “魔偶上传 3/3”. Mini sequence strip “1C 55 BD E9 7A 55 FF”, technical small seed “种子 NB-7C9F”. Two outline actions “查看复盘” and “以此种子练习”. Bottom-left “Esc 返回”; bottom-right understated “导出记录”. No global online leaderboard, no avatars. Professional Cyberpunk inventory-list visual language. Keep all typography clear.
```

## 06-settings

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: full-screen authentic game SETTINGS, not a desktop preferences dialog. Header “设置” yellow at top-left; tabs along top “音频” selected cyan, “画面”, “操作”, “语言与辅助”. Thin coral-red separator. Main centered wide settings column with left labels, right numeric controls, segmented red tracks with cyan current fill, tiny angular arrow buttons:
section “音量”
“总音量” 80
“音乐音量” 65
“音效音量” 85
section “音乐行为”
“动态音乐” “开启”
“后台静音” “开启”
section “故障视觉效果”
“故障效果强度” “低”
“减少动态效果” “关闭”
“低时间闪烁” “关闭”
Right narrow explanatory pane titled “动态音乐” with text “随入侵阶段变化。最后 10 秒增强节奏，暂停时淡出。” Under this a tiny tasteful waveform and “工业电子 / 合成器氛围”.
Bottom red text “M 静音”; footer buttons “试听音乐” red outline, “恢复默认” red outline, “应用” bright yellow, “Esc 返回”.
Across upper-right margin ONE tiny brief glitch frozen in time: cyan/red RGB split fragments, extremely restrained. Must keep labels clear and avoid row overlaps. Main UI is charcoal-black with red rules, exact functional sliders, flat and spacious like Cyberpunk 2077 settings.
```

## 07-pause

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: full-screen PAUSE menu in an active hacking game. Top-left small yellow game wordmark. Large yellow heading “连接已暂停”. Left center vertical menu with five red items: “继续游戏”, “保存并返回主菜单”, “设置”, “重新开始”, “退出游戏”. First selected in translucent cyan horizontal bar with cyan angular border, same look as Cyberpunk main menu. Right area contains NO legible underlying puzzle: completely opaque near-black panel with faint red circuit diagram, blacked-out code board for pause fairness. Right small status only “标准挑战 · 标准”, “剩余 42.00 秒”, “缓冲区 3/8”, cyan small save status “进度已保存”. At bottom understated explanation “计时已停止，恢复后继续当前入侵。” and keyboard hints “Esc 继续    Enter 确认”. A dim industrial terminal background is acceptable, but no matrix tokens visible, no targets visible. No browser or window chrome. Add only a tiny frozen scarlet/cyan horizontal glitch fragment near header, not over letters.
```

## 08-tutorial

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
SCREEN: interactive NEW PLAYER TUTORIAL, game UI rather than infographic poster. Large yellow title “入侵训练”; top-right “第 2 步 / 共 4 步”. Same dark red terminal borders as game. Left 60% a large simple 4x4 practice code matrix, exactly four rows of four cells:
1C 55 BD E9
7A E9 55 1C
BD 7A FF 55
FF BD 1C 7A
Only row1 col1 “1C” is already selected, dim with tiny step1. Column1 remaining cells highlighted pale yellow vertical stripe. row3 col1 “BD” has cyan focus brackets. Above matrix buffer five outlined slots, first “1C” and four empty. Right 34% a bright yellow heading “接下来选同一列”, ivory short sentence “第一次从第一行选码。之后在列和行之间交替。” Below a target example “1C BD 7A”, first 1C cyan, next BD yellow. Instruction “选择第 3 行的 BD” with small pointer cue toward selected column. Bottom-right yellow action “继续练习”; left small “上一步” and “退出训练”. Bottom fine note “训练不限时，不计入最佳成绩”. Legible Chinese, strong didactic spacing, distinct active column, authentic hacking game atmosphere.
```

## 09-state-variants

```text
Use case: ui-mockup. Create a polished production-quality 16:9 desktop videogame UI concept screenshot, ideally 1920x1080 or higher. It must closely recreate the visual language of Cyberpunk 2077's ACTUAL in-game menus and Breach Protocol interface: near-black burgundy #100C10 backgrounds, angular thin coral-red #F25B64 linework and red typography, toxic yellow #FCEE09 headings and primary actions, luminous pale cyan #75F0E7 selection bars and confirmed codes, ivory #E6DDE1 readable body copy, dim muted red secondary text. NOT generic futuristic SaaS, NOT purple neon, NOT rounded card dashboard, NOT a web page. Flat front-on screen, edge-to-edge game viewport, no device frame, no macOS traffic lights. Severe cut corners, red registration ticks, sparse technical circuit silhouettes, carefully spaced dense functional HUD, very restrained scanlines, absolutely legible labels. Sharp condensed futuristic sans headings and monospaced two-character hex codes, Chinese simplified sans labels. Actual usable full-screen PC game visual hierarchy. Brand is original text “神经裂隙” with small English “NEURAL BREACH” in a slanted jagged yellow wordmark reminiscent of the game's typography, no CDPR or Cyberpunk logo. Shared identity across all pages. No website navigation, no browser chrome, no watermark. Text in Chinese must be accurate and clear. Any tiny secondary text may be minimal.
ASSET: a premium UI design comparison sheet containing TWO equal full 16:9 game screen mockups stacked vertically, clearly separated by a thin red gap, canvas approximately 16:18. Both must have legible sizable Chinese typography and same palette.
TOP SCREEN: failed breach result. Top yellow title “入侵结束”, large coral-red headline “连接已断开”. Left giant “0” label “本局得分”, small “上传 0/3”. Right flat red-bordered failure explanation “时间耗尽” and “本次未完成任何魔偶序列。” then guidance “复盘路径，尝试利用序列重叠。” Bottom buttons yellow “再来一局”, red outline “查看复盘”, red text “返回主菜单”. Red/cyan shredded digital glitch fragments around outer edges only, never text. No huge warning triangle.
BOTTOM SCREEN: save-and-exit confirmation modal centered over fully darkened paused game. Angular black red-outline dialog width60%, yellow title “退出游戏？”. Body “当前入侵进度将保存。下次可从主菜单继续。” Small cyan “标准 · 剩余 42.00 秒 · 缓冲区 3/8”. Two large bottom buttons: red outline “取消”, yellow “保存并退出”. Background obscures all matrix codes. Include a small outer sheet caption “失败结算 / 退出确认” at top in muted red. These are functional videogame screens, not web dialogs.
```

## 03-breach-game 局部修订

```text
Edit this game UI screenshot with one precise correction ONLY. Preserve every pixel of the rest of the interface, the entire 6x6 matrix, targets, typography and colors. Top-right buffer currently wrongly has NINE slots. Redraw ONLY the interior buffer slot row as EXACTLY EIGHT equal rectangular slots TOTAL: first 1C, second 55, third BD, fourth EMPTY, fifth EMPTY, sixth EMPTY, seventh EMPTY, eighth EMPTY. Thus THREE filled + FIVE empty = EIGHT. Keep the heading 缓冲区 3/8. Use wider slots if necessary to occupy the existing row. Do not add a ninth slot. Everything else remains unchanged.
```

## 09-state-variants 局部修订

```text
Edit this two-panel game UI state sheet. Preserve top failure panel, centered exit confirmation dialog, all buttons, headings and layout. Change ONLY the dimmed background of the LOWER screen outside the exit dialog: replace the entire underlying code matrix, buffer, targets and all their alphanumeric letters with an OPAQUE black-burgundy background (#100C10), maybe faint abstract red circuit traces but NO readable grid, numbers, codes, targets, or buffer at all. This must fully hide the paused puzzle for game fairness, not just dim it. Keep the dialog text and its 42.00, 3/8 status unchanged. No other changes.
```
