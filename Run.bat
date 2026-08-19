@echo off
setlocal EnableExtensions
call "%~dp0scripts\windows\Run_impl.bat" %*
exit /b %errorlevel%
