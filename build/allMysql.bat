@echo off
REM Init MySQL schema for ALL server repos (7): core + ai/backend/call/online/open/social
REM Serial: DB init is a one-shot task, exits when done (no idle windows).
setlocal enabledelayedexpansion

for %%i in ("%~dp0..\..") do set "WORKSPACE=%%~fi"

echo ============================================
echo   Init all databases (7 repos)
echo   Workspace: !WORKSPACE!
echo ============================================

set "FAIL=0"

call :init "Core beaver-server"     "!WORKSPACE!\beaver-server"                               "go run main.go -db"
if errorlevel 1 set /a FAIL+=1
call :init "AI beaver_agent"        "!WORKSPACE!\beaver-server-package\beaver-server-ai"      ".venv\Scripts\python.exe main.py -db"
if errorlevel 1 set /a FAIL+=1
call :init "Backend beaver_backend" "!WORKSPACE!\beaver-server-package\beaver-server-backend"  "go run main.go -db"
if errorlevel 1 set /a FAIL+=1
call :init "Call beaver_call"       "!WORKSPACE!\beaver-server-package\beaver-server-call"     "go run main.go -db"
if errorlevel 1 set /a FAIL+=1
call :init "Office beaver_office"   "!WORKSPACE!\beaver-server-package\beaver-server-online"   "go run main.go -db"
if errorlevel 1 set /a FAIL+=1
call :init "Open beaver_open"       "!WORKSPACE!\beaver-server-package\beaver-server-open"     "go run main.go -db"
if errorlevel 1 set /a FAIL+=1
call :init "Social circle/moment"   "!WORKSPACE!\beaver-server-package\beaver-server-social"   "go run main.go -db"
if errorlevel 1 set /a FAIL+=1

echo ============================================
if "!FAIL!"=="0" (
    echo   [OK] All databases initialized
) else (
    echo   [FAIL] !FAIL! repos failed - see errors above
)
echo ============================================
set "RC=!FAIL!"
endlocal & exit /b %RC%

:init
setlocal
set "LABEL=%~1"
set "DIR=%~2"
set "CMD=%~3"
echo.
echo ------ !LABEL! ------
if not exist "!DIR!" (
    echo   [FAIL] dir not found: !DIR!
    endlocal & exit /b 1
)
pushd "!DIR!" || (endlocal & exit /b 1)
!CMD!
set "RC=!ERRORLEVEL!"
popd
if "!RC!"=="0" (
    echo   [OK] !LABEL!
) else (
    echo   [FAIL] !LABEL! ^(exit=!RC!^)
)
endlocal & exit /b !RC!
