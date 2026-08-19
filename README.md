# PromptNest

> 一个轻量、精致的本地 Prompt 管理器，使用 **Go + Wails + Vue 3** 开发。

**作者 / GitHub：** https://github.com/leon-dev-lab  
**计划仓库：** https://github.com/leon-dev-lab/PromptNest  
**许可证：** MIT

PromptNest 使用自定义 Warm Editorial Workspace 风格界面，不依赖大型 UI 组件库。数据默认保存在本地，支持自定义数据目录。

## 功能

- 提示词新建、编辑、删除、复制、收藏、创建副本
- 标题 / 内容 / 分类 / 标签 / 备注全文搜索
- 提示词自由拖拽排序并持久化
- 分类新增、删除、重命名、拖拽排序、20 色自定义颜色
- Tag Chips 标签输入：输入后按 Enter 创建，点击 × 删除
- 基于当前有效标签库的历史标签联想
- 650ms 防抖自动保存
- JSON 导入 / 导出
- 自定义数据目录
- 浅色 / 深色模式
- 系统托盘、关闭到托盘、双击恢复
- Windows x64 构建脚本
- 软件内直接打开 `leon-dev-lab` GitHub 主页

## 首次启动示例

首次运行且本地还没有 `prompts.json` 时，会自动创建演示数据。

示例分类：

- 通用分类
- 写作
- 开发
- 办公
- 产品
- 分析
- 图像

示例提示词：

- 结构化任务助手
- 文章润色
- 代码审查助手
- 会议纪要整理
- 需求拆解
- 表格数据洞察
- 图片内容命名

这些只是首次启动示例，可以自由修改或全部删除。已有用户数据不会被默认示例覆盖。

## 技术栈

- Go
- Wails v2.14.0
- Vue 3
- Vite
- Windows WebView2 Runtime
- 自定义 CSS UI

## Windows 开发环境

建议准备：

```text
Go
Node.js + npm
WebView2 Runtime
```

项目不要求全局安装 Wails CLI，构建脚本会使用固定 Wails v2.14.0。

## 一键运行

```bat
Run.bat
```

如果尚未构建，会自动先构建再启动。

## 构建 Windows x64

```bat
Build_EXE.bat
```

输出：

```text
build\bin\PromptNest.exe
```

## 生成 GitHub Release ZIP

```bat
Package_Release.bat
```

成功后会生成：

```text
release\PromptNest-v2.6.0-win-x64.zip
```

可以直接上传到 GitHub Releases。

## 默认数据位置

Windows：

```text
%LOCALAPPDATA%\PromptNest\prompts.json
```

应用内可以切换到任意自定义数据目录。

> 仓库本身不会包含你的个人 `prompts.json`、`settings.json`、API Key 或日志文件。

## 开源许可

PromptNest 使用 [MIT License](LICENSE)。
