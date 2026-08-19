# Windows scripts

这些脚本是 PromptNest 的 Windows 开发与发布辅助工具。

- `Build_EXE_impl.bat`：根目录 `Build_EXE.bat` 的实际构建实现。
- `Run_impl.bat`：根目录 `Run.bat` 的实际启动实现。
- `Build_And_Run.bat`：强制重新构建并启动。
- `Dev.bat`：Wails 开发模式。
- `Check_Environment.bat`：检查 Go / Node / npm / GOPROXY。
- `Package_Release.bat`：生成 Windows x64 Release ZIP。
- `Set_GoProxy_CN.bat`：将当前 Windows 用户的 Go Proxy 设置为 `goproxy.cn`。

普通用户只需要使用仓库根目录的 `Run.bat`；需要构建 Windows x64 时使用根目录的 `Build_EXE.bat`。
