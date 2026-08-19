@echo off
setlocal EnableExtensions
for %%I in ("%~dp0..\..") do set "ROOT=%%~fI"
cd /d "%ROOT%"
echo ================================================
echo  PromptNest - Environment Check
echo ================================================
echo.

echo [Go]
where go
if errorlevel 1 (echo NOT FOUND) else go version
echo.

echo [Node]
where node
if errorlevel 1 (echo NOT FOUND) else node -v
echo.

echo [npm]
where npm.cmd
set "NPM_VER="
for /f "delims=" %%V in ('call npm.cmd -v 2^>nul') do set "NPM_VER=%%V"
if defined NPM_VER (echo npm %NPM_VER%) else echo npm NOT WORKING - npm -v returned no output
echo.

echo [Go proxy]
go env GOPROXY 2>nul
echo.
echo Wails CLI does NOT need to be installed globally.
echo Build scripts use: go run github.com/wailsapp/wails/v2/cmd/wails@v2.14.0
echo.
pause
