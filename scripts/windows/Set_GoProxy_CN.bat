@echo off
setlocal EnableExtensions
for %%I in ("%~dp0..\..") do set "ROOT=%%~fI"
cd /d "%ROOT%"
echo This sets GOPROXY globally for your current Windows user:
echo   https://goproxy.cn,direct
echo.
go env -w GOPROXY=https://goproxy.cn,direct
if errorlevel 1 (
  echo [ERROR] Failed to set GOPROXY.
) else (
  echo Done. Current GOPROXY:
  go env GOPROXY
)
echo.
pause
