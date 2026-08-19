@echo off
setlocal EnableExtensions
for %%I in ("%~dp0..\..") do set "ROOT=%%~fI"
cd /d "%ROOT%"
echo This will rebuild PromptNest and launch it automatically.
echo.
call "%ROOT%\Build_EXE.bat" /run
exit /b %errorlevel%
