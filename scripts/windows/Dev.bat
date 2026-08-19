@echo off
setlocal EnableExtensions
for %%I in ("%~dp0..\..") do set "ROOT=%%~fI"
cd /d "%ROOT%"
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

pushd "%ROOT%\frontend"
if not exist node_modules (
  echo Installing frontend dependencies...
  call npm.cmd install
  if errorlevel 1 (popd & echo [ERROR] npm install failed.& pause & exit /b 1)
)
popd

rem main.go embeds frontend/dist. A fresh clone may not have generated frontend output yet.
if not exist "%ROOT%\frontend\dist" mkdir "%ROOT%\frontend\dist" >nul 2>nul
if not exist "%ROOT%\frontend\dist\index.html" (
  > "%ROOT%\frontend\dist\index.html" echo ^<!doctype html^>^<html^>^<body^>dev placeholder^</body^>^</html^>
)

echo Starting Wails dev mode. The desktop window should open automatically after compilation.
go run github.com/wailsapp/wails/v2/cmd/wails@%WAILS_VERSION% dev
if errorlevel 1 (
  echo.
  echo [ERROR] Wails dev exited with an error.
  pause
  exit /b 1
)
