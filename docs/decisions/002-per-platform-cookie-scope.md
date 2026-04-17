# 002 — Cookie 按平台分文件存储

## 背景

ali-mcp 同时服务淘宝和闲鱼。两者共享阿里账号体系（SSO），但 cookie 的 domain scope 不同：

- 淘宝站：`.taobao.com`、`.tmall.com`
- 闲鱼 h5：`.goofish.com`、`.m.goofish.com`

参考项目 `xiaohongshu-mcp` 只服务单平台，cookie 存单个 `cookies.json` 即可。

## 选项

1. **单文件**：所有域的 cookie 混存到 `cookies.json`，加载时不过滤
2. **按 domain 过滤单文件**：存单文件，加载时按当前访问的域做白名单过滤
3. **按平台分文件**：`cookies/taobao.json` + `cookies/xianyu.json`，访问前选对应文件

## 决策

选 **3：按平台分文件**。

## 理由

- **隔离清晰**：淘宝登录失效不会波及闲鱼（虽然账号体系相同，但 cookie 过期时间独立）
- **调试友好**：手动删除某平台 cookie 可精确重置登录态
- **实现简单**：`browser.WithPlatform(p)` 就自动选对文件，不需要每次判断 domain
- **fail loud**：`cookies.GetCookiesFilePath(invalid)` 直接 panic，防止 cookie 串流
- **可扩展**：后续加 alimama / 1688 等平台，直接加常量即可

## 代价

- SSO 回流场景：用户扫码登淘宝后，需要再跑一次 "访问闲鱼 h5 → 保存闲鱼 cookie" 才能让闲鱼模块工作。首次登录流程会多一步。
