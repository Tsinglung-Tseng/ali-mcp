# 闲鱼发布功能 - 实机研究指南

> Created: 2026-04-27
> Status: PENDING — 等待人工 headful 走一遍并填回选择器
> 关联代码：`xianyu/publish.go` (stub)

## 目的

`xianyu_publish` 这个 MCP 工具不能闭门盲写：发布是 5 步表单 + 文件上传 + 可能的风控验证，每一步都需要实机捕获 DOM 才能写出能跑的代码。本文档列出走查清单，跑一遍后即可把选择器填回 `xianyu/publish.go`。

## 准备

```bash
cd ~/scaffold/ali.mcp
./ali-mcp -headless=false -port=:18070
```

确保已通过 `taobao_get_login_qrcode` + 访问 `h5.m.goofish.com` 触发 SSO 拿到闲鱼 cookie。

## 走查清单

### 步骤 1：发布入口

- [ ] 打开 `https://h5.m.goofish.com`，找"发布"按钮
- [ ] 记录入口 URL（点击后跳到的发布页 URL，如 `https://h5.m.goofish.com/publish.html`）
- [ ] 注意：h5 站可能没有完整发布入口，需走 m.goofish.com / 主站 / app

### 步骤 2：图片上传

- [ ] DevTools 找上传按钮：是 `<input type="file">` 还是覆盖在 `<div>` 上的隐藏 input？
- [ ] 记录选择器：`<input type="file">` 的精确 selector
- [ ] 记录是否多选（`multiple` 属性）
- [ ] 注意单图最大尺寸 / 张数限制（写到 PublishArgs 注释里）
- [ ] 上传后等待时间：通常需要等图片上传完成才能继续，记录 loading 指示器选择器（用于 `WaitFor`）

### 步骤 3：标题 + 描述

- [ ] 标题输入框 selector
- [ ] 描述（textarea）selector
- [ ] 字符数限制（如 30/500），写到 PublishArgs 注释

### 步骤 4：分类

- [ ] 分类选择是模态弹窗还是嵌入式选择器？
- [ ] 是否有"建议分类"自动填？
- [ ] 是否能用纯文本搜索 + 选首个匹配？
- [ ] 记录每一级分类的选择器：`.cat-level-1 [data-name="X"]` 这类
- [ ] 完成后如何关闭（确认按钮 selector）

### 步骤 5：价格

- [ ] 价格输入框 selector
- [ ] 原价（划线价）输入框 selector
- [ ] 是否有"包邮"/运费输入

### 步骤 6：定位

- [ ] 默认是不是用浏览器定位？headless 下没有定位权限会发生什么？
- [ ] 是否能手动输入城市
- [ ] 城市选择器 selector

### 步骤 7：提交

- [ ] 提交按钮 selector
- [ ] 提交后跳转到哪个 URL（用于成功判断）
- [ ] 失败提示 selector + 文案样本

### 步骤 8：风控

- [ ] headful 模式提交会不会出滑块？
- [ ] headless 模式呢？
- [ ] 如果出了，记录滑块容器选择器，决定是否需要 stealth 增强

## 反馈到代码

跑完后，编辑 `xianyu/publish.go`：

1. 把 `Publish` 的 stub 实现替换成真实流程
2. 顶部加 `const (...)` 块装所有选择器，命名按 `pubXxxSel`
3. 每步前后加 `time.Sleep` 让动画 / 网络完成
4. 失败路径走 `errors.Wrap`

## 已知坑（猜测）

- **CSS class hash**：闲鱼 h5 是 React 组件，所有 class 带 hash，必须用前缀匹配
- **拖拽上传**：如果不是 `<input type="file">` 而是拖拽组件，要用 `Page.AddScriptTag` 注入文件或用 CDP 的 `Input.dispatchDragEvent`
- **风控 cookie**：长时间未发过的账号一发就被风控，建议先小成本测试（一个真要卖的便宜物品）
- **图片预处理**：闲鱼对图片尺寸/比例有限制，最好预先压成 1080×1080 以下

## 完成定义

跑完一次成功发布（哪怕是一个 1 元测试品并立即下架），把过程录屏 / 关键截图存到 `docs/debug/xianyu-publish-screenshots/`，把选择器和坑都回填到本文档。
