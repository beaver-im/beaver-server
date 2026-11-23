@echo off
REM 初始化 beaver-server 核心库表（beaver_account/relation/message/asset/notification/platform/datasync/agent）
setlocal enabledelayedexpansion

for %%i in ("%~dp0..") do set "ROOT_DIR=%%~fi"

wt new-tab --title "Beaver Core DB Init" cmd /k "cd /d !ROOT_DIR! && go run main.go -db"
