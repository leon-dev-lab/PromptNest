@echo off
setlocal EnableExtensions
for %%I in ("%~dp0..\..") do set "ROOT=%%~fI"
cd /d "%ROOT%"

set "WAILS_VERSION=v2.14.0"
set "APP_EXE=%ROOT%\build\bin\PromptNest.exe"
set "RUN_AFTER_BUILD=0"
if /I "%~1"=="/run" set "RUN_AFTER_BUILD=1"

echo ================================================
echo  PromptNest 2.6.0 - Production Build
echo  Go + Wails %WAILS_VERSION% + Vue 3
echo ================================================
echo.

echo [1/6] Checking Go...
where go >nul 2>nul
if errorlevel 1 goto :no_go
for /f "delims=" %%V in ('go version 2^>nul') do set "GO_VER=%%V"
if not defined GO_VER goto :no_go
echo       %GO_VER%

echo [2/6] Checking Node and npm...
where node >nul 2>nul
if errorlevel 1 goto :no_node
where npm.cmd >nul 2>nul
if errorlevel 1 goto :no_npm
for /f "delims=" %%V in ('node -v 2^>nul') do set "NODE_VER=%%V"
for /f "delims=" %%V in ('call npm.cmd -v 2^>nul') do set "NPM_VER=%%V"
if not defined NPM_VER goto :broken_npm
echo       Node %NODE_VER% / npm %NPM_VER%

echo [3/6] Restoring Go modules...
go mod download
if errorlevel 1 (
    echo       Default Go proxy failed. Retrying with goproxy.cn for this build only...
    set "GOPROXY=https://goproxy.cn,direct"
    go mod download
)
if errorlevel 1 goto :go_modules_failed

echo [4/6] Installing frontend dependencies...
pushd "%ROOT%\frontend"
if exist package-lock.json (
    call npm.cmd ci
) else (
    call npm.cmd install
)
if errorlevel 1 (
    popd
    goto :npm_failed
)
popd

echo [5/6] Building Windows x64 EXE...
echo       Cleaning stale frontend/build output...
if exist "%ROOT%\frontend\dist" rmdir /s /q "%ROOT%\frontend\dist" >nul 2>nul
mkdir "%ROOT%\frontend\dist" >nul 2>nul
> "%ROOT%\frontend\dist\index.html" echo ^<!doctype html^>^<html^>^<body^>build placeholder^</body^>^</html^>
if exist "%ROOT%\build\bin" rmdir /s /q "%ROOT%\build\bin" >nul 2>nul
go run github.com/wailsapp/wails/v2/cmd/wails@%WAILS_VERSION% build -clean -platform windows/amd64 -o PromptNest.exe
if errorlevel 1 goto :wails_failed

echo [6/6] Verifying output...
if not exist "%APP_EXE%" goto :exe_missing

echo.
echo ================================================
echo  BUILD SUCCESS
echo ================================================
echo  %APP_EXE%
echo.

if "%RUN_AFTER_BUILD%"=="1" (
    echo Launching PromptNest.exe ...
    start "" /D "%ROOT%\build\bin" "%APP_EXE%"
    if errorlevel 1 goto :launch_failed
)
exit /b 0

:no_go
echo [ERROR] Go is not available in PATH.
goto :failed_pause
:no_node
echo [ERROR] Node.js is not available in PATH.
goto :failed_pause
:no_npm
echo [ERROR] npm.cmd was not found. Reinstall Node.js with npm enabled.
goto :failed_pause
:broken_npm
echo [ERROR] npm was found, but "npm -v" returned no version.
echo         Your npm installation is incomplete or broken.
goto :failed_pause
:go_modules_failed
echo [ERROR] Unable to download Go modules.
echo         Check network access or GOPROXY.
goto :failed_pause
:npm_failed
echo [ERROR] npm dependency installation failed.
goto :failed_pause
:wails_failed
echo [ERROR] Wails production build failed.
goto :failed_pause
:exe_missing
echo [ERROR] Build command finished but EXE was not found:
echo         %APP_EXE%
goto :failed_pause
:launch_failed
echo [ERROR] EXE was built but Windows failed to start it.
goto :failed_pause
:failed_pause
echo.
pause
exit /b 1
