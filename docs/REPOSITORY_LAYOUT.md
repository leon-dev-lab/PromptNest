# Repository layout notes

本次仅整理仓库结构，不改变 PromptNest 的业务功能、数据格式或 Wails 前端结构。

## 根目录保留

根目录继续保留：

- `Run.bat`：普通 Windows 用户的一键运行入口。
- `Build_EXE.bat`：Windows x64 一键构建入口。
- 核心 Go 源码：`main.go`、`app.go`、`datastore.go`、`models.go`、`tray_*.go` 及对应测试。
- `README.md`、`LICENSE`、`go.mod`、`wails.json` 等标准项目文件。

Go 源码没有为了“目录好看”而强行拆包，避免引入没有必要的 package/import 重构风险。

## 移动的文件

以下 Windows 辅助脚本移动到 `scripts/windows/`：

- `Build_And_Run.bat`
- `Check_Environment.bat`
- `Dev.bat`
- `Package_Release.bat`
- `Set_GoProxy_CN.bat`

`Build_EXE.bat` 与 `Run.bat` 仍位于根目录，但现在是稳定入口包装器，实际实现位于 `scripts/windows/`。

文档调整：

- `CONTRIBUTING.md` → `.github/CONTRIBUTING.md`
- `CHANGELOG.md` → `docs/CHANGELOG.md`
- `RELEASE_NOTES_v2.6.0.md` → `docs/releases/v2.6.0.md`

## 迁移现有 Git 仓库

如果你把整理后的文件复制到现有仓库，请确认旧根目录中的上述文件已经删除，否则 GitHub 页面仍会同时显示旧文件和新目录中的文件。

推荐使用 GitHub Desktop 检查本次变更：旧位置应显示为删除，新位置应显示为新增/重命名，然后再提交。
