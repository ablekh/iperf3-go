@echo off
echo Building iperf3-go server...
go build -o iperf3-go.exe main.go

if %errorlevel% neq 0 (
    echo Build failed!
    exit /b 1
)

echo Starting iperf3-go server in background...
start /b iperf3-go.exe -v

timeout /t 2 /nobreak >nul

echo Testing with iperf3 client...
echo Running: iperf3 -c localhost -t 5

where iperf3 >nul 2>&1
if %errorlevel% equ 0 (
    iperf3 -c localhost -t 5
    set TEST_RESULT=%errorlevel%
) else (
    echo iperf3 client not found. Please install iperf3 to test:
    echo   Download from https://iperf.fr/iperf-download.php
    echo   Or use chocolatey: choco install iperf3
    set TEST_RESULT=1
)

echo Stopping server...
taskkill /f /im iperf3-go.exe >nul 2>&1

if %TEST_RESULT% equ 0 (
    echo Test completed successfully!
) else (
    echo Test failed or iperf3 client not available
)

exit /b %TEST_RESULT%