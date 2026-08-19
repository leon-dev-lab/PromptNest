@echo off
setlocal EnableExtensions
cd /d "%~dp0"
echo This will rebuild PromptNest and launch it automatically.
echo.
call "%~dp0Build_EXE.bat" /run
if errorlevel 1 exit /b 1
exit /b 0
