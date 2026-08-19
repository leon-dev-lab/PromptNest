@echo off
setlocal EnableExtensions
cd /d "%~dp0"
set "WAILS_VERSION=v2.14.0"

echo ================================================
echo  PromptNest - Wails Development Mode
echo ================================================
where go >nul 2>nul || (echo [ERROR] Go not found.& pause & exit /b 1)
where npm.cmd >nul 2>nul || (echo [ERROR] npm.cmd not found.& pause & exit /b 1)

for /f "delims=" %%V in ('call npm.cmd -v 2^>nul') do set "NPM_VER=%%V"
if not defined NPM_VER (
  echo [ERROR] npm exists but did not return a version.
  pause
  exit /b 1
)

pushd frontend
if not exist node_modules (
  echo Installing frontend dependencies...
  call npm.cmd install
  if errorlevel 1 (popd & echo [ERROR] npm install failed.& pause & exit /b 1)
)
popd

rem main.go embeds frontend/dist. A fresh Git clone may not have the ignored dist directory yet.
if not exist "%~dp0frontend\dist" mkdir "%~dp0frontend\dist" >nul 2>nul
if not exist "%~dp0frontend\dist\index.html" (
  > "%~dp0frontend\dist\index.html" echo ^<!doctype html^>^<html^>^<body^>dev placeholder^</body^>^</html^>
)

echo Starting Wails dev mode. The desktop window should open automatically after compilation.
go run github.com/wailsapp/wails/v2/cmd/wails@%WAILS_VERSION% dev
if errorlevel 1 (
  echo.
  echo [ERROR] Wails dev exited with an error.
  pause
  exit /b 1
)
