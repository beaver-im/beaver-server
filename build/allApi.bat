@echo off
REM Start ALL API services (no RPC; RPC is started by each repo start_rpc.bat)
REM Core 10 APIs + 6 business repos x1 = 16 tabs.
setlocal enabledelayedexpansion

for %%i in ("%~dp0..\..") do set "WORKSPACE=%%~fi"

echo Workspace: !WORKSPACE!
echo Starting all API services (16)...

wt new-tab --title "Gateway API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\gateway\gateway_api && go run gateway.go" ^
  ; new-tab --title "Account API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\account\account_api && go run account.go" ^
  ; new-tab --title "Asset API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\asset\asset_api && go run asset.go" ^
  ; new-tab --title "Relation API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\relation\relation_api && go run relation.go" ^
  ; new-tab --title "Message API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\message\message_api && go run message.go" ^
  ; new-tab --title "Platform API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\platform\platform_api && go run platform.go" ^
  ; new-tab --title "Notification API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\notification\notification_api && go run notification.go" ^
  ; new-tab --title "Push API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\push\push_api && go run push.go" ^
  ; new-tab --title "Datasync API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\datasync\datasync_api && go run datasync.go" ^
  ; new-tab --title "WS API" cmd /k "cd /d !WORKSPACE!\beaver-server\app\ws\ws_api && go run ws.go" ^
  ; new-tab --title "Agent API (AI)" cmd /k "cd /d !WORKSPACE!\beaver-server-package\beaver-server-ai && .venv\Scripts\python.exe app\ai\agent_api\main.py" ^
  ; new-tab --title "Backend Admin" cmd /k "cd /d !WORKSPACE!\beaver-server-package\beaver-server-backend\app\backend_admin && go run backend.go" ^
  ; new-tab --title "Call API" cmd /k "cd /d !WORKSPACE!\beaver-server-package\beaver-server-call\app\call_api && go run call.go" ^
  ; new-tab --title "Office API" cmd /k "cd /d !WORKSPACE!\beaver-server-package\beaver-server-online\app\office\office_api && go run office.go" ^
  ; new-tab --title "Open API" cmd /k "cd /d !WORKSPACE!\beaver-server-package\beaver-server-open\app\open\open_api && go run open.go" ^
  ; new-tab --title "Social API" cmd /k "cd /d !WORKSPACE!\beaver-server-package\beaver-server-social\app\social\social_api && go run social.go"

endlocal
