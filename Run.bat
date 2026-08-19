@echo off
setlocal EnableExtensions
cd /d "%~dp0"
set "APP_EXE=%~dp0build\bin\PromptNest.exe"

echo ================================================
echo  PromptNest - One Click Run
echo ================================================

if exist "%APP_EXE%" (
    echo Found existing EXE. Launching now...
    start "" /D "%~dp0build\bin" "%APP_EXE%"
    if errorlevel 1 (
        echo [ERROR] Windows could not launch the EXE.
        pause
        exit /b 1
    )
    exit /b 0
)

echo EXE not found. Building automatically first...
echo.
call "%~dp0Build_EXE.bat" /run
if errorlevel 1 (
    echo.
    echo Build or launch failed. See the error above.
    pause
    exit /b 1
)
exit /b 0
