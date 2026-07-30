@echo off
chcp 65001 >nul

cd /d "%~dp0"

start "xcgui-shadcn-ui frontend" cmd /k dev_start_frontend.bat

start "xcgui-shadcn-ui backend" cmd /k dev_start_backend.bat

exit