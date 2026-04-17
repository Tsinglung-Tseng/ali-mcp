# 001 — Go Module Path

## 背景

新建仓库需要决定 Go module 路径。参考项目 `xiaohongshu-mcp` 用 `github.com/xpzouying/xiaohongshu-mcp`（连字符），但本项目目录名是 `ali.mcp`（点号）。

## 选项

1. `ali-mcp` — 纯本地名，无仓库前缀
2. `github.com/Tsinglung-Tseng/ali-mcp` — 连字符，符合 Go 惯例
3. `github.com/Tsinglung-Tseng/ali.mcp` — 点号，保持目录名一致

## 决策

选 **3：`github.com/Tsinglung-Tseng/ali.mcp`**。

## 理由

- 确认要推 GitHub，带仓库前缀才能被 `go get` 消费
- 保留点号匹配目录名，避免将来 clone 时 "目录名和 import 路径不一致" 的认知负担
- Go 允许 module 路径含点号；仅限制不能用 `+` / 某些保留字
- 迁移代价低：import 就一处改动，后续真要改成连字符可以一把 sed
