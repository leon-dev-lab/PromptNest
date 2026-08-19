@echo off
setlocal EnableExtensions
call "%~dp0scripts\windows\Build_EXE_impl.bat" %*
exit /b %errorlevel%
