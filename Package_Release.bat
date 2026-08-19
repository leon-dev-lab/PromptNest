@echo off
setlocal EnableExtensions
cd /d "%~dp0"

set "VERSION=2.6.0"
set "APP_EXE=%~dp0build\bin\PromptNest.exe"
set "RELEASE_ROOT=%~dp0release"
set "STAGE=%RELEASE_ROOT%\PromptNest-v%VERSION%-win-x64"
set "ZIP_FILE=%RELEASE_ROOT%\PromptNest-v%VERSION%-win-x64.zip"

echo ================================================
echo  PromptNest v%VERSION% - GitHub Release Package
echo ================================================
echo.

call "%~dp0Build_EXE.bat"
if errorlevel 1 goto :failed
if not exist "%APP_EXE%" goto :missing

if exist "%STAGE%" rmdir /s /q "%STAGE%"
if exist "%ZIP_FILE%" del /q "%ZIP_FILE%"
mkdir "%STAGE%" >nul 2>nul

copy /y "%APP_EXE%" "%STAGE%\PromptNest.exe" >nul
copy /y "%~dp0LICENSE" "%STAGE%\LICENSE.txt" >nul

> "%STAGE%\README.txt" echo PromptNest v%VERSION%
>>"%STAGE%\README.txt" echo.
>>"%STAGE%\README.txt" echo GitHub: https://github.com/leon-dev-lab
>>"%STAGE%\README.txt" echo Repository: https://github.com/leon-dev-lab/PromptNest
>>"%STAGE%\README.txt" echo.
>>"%STAGE%\README.txt" echo Run PromptNest.exe directly. Windows WebView2 Runtime is required.
>>"%STAGE%\README.txt" echo User data is stored outside this release folder by default.

powershell -NoProfile -ExecutionPolicy Bypass -Command "Compress-Archive -Path '%STAGE%\*' -DestinationPath '%ZIP_FILE%' -CompressionLevel Optimal -Force"
if errorlevel 1 goto :zipfailed

echo.
echo ================================================
echo  RELEASE PACKAGE READY
echo ================================================
echo  %ZIP_FILE%
echo.
exit /b 0

:missing
echo [ERROR] PromptNest.exe was not found after build.
goto :failed
:zipfailed
echo [ERROR] Failed to create release ZIP.
goto :failed
:failed
echo.
pause
exit /b 1
