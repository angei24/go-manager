@echo off
REM gm installer launcher for Windows CMD
REM Usage: install.bat [--from-source] [--add-to-path]

setlocal
set "SCRIPT_DIR=%~dp0"
powershell -NoProfile -ExecutionPolicy Bypass -File "%SCRIPT_DIR%install.ps1" %*
exit /b %ERRORLEVEL%
