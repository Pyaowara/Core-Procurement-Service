@echo off
setlocal enabledelayedexpansion
cd /d %~dp0

cls
echo.
echo ============================================================
echo   Core Procurement Service - Windows Startup Script
echo ============================================================
echo.
echo Starting all microservices in separate terminals...
echo.

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed or not in PATH!
    echo Please install Go from https://golang.org/
    pause
    exit /b 1
)

REM Start each service
echo [1/4] Starting auth-identity-service...
start "auth-identity-service" cmd /k "cd /d %~dp0services\auth-identity-service && echo Service: auth-identity-service && echo. && go run main.go"
timeout /t 1 /nobreak >nul

echo [2/4] Starting inventory-service...
start "inventory-service" cmd /k "cd /d %~dp0services\inventory-service && echo Service: inventory-service && echo. && go run main.go"
timeout /t 1 /nobreak >nul

echo [3/4] Starting purchase-service...
start "purchase-service" cmd /k "cd /d %~dp0services\purchase-service && echo Service: purchase-service && echo. && go run main.go"
timeout /t 1 /nobreak >nul

echo [4/4] Starting approval-service...
start "approval-service" cmd /k "cd /d %~dp0services\approval-service && echo Service: approval-service && echo. && go run main.go"

echo.
echo ============================================================
echo All services are starting up!
echo Check the individual terminal windows for service status.
echo ============================================================
echo.
echo Press any key to close this window...
pause >nul
exit /b 0
