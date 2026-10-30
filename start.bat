@echo off
setlocal
cd /d "%~dp0"

echo [1/3] Close Old Service...
taskkill /f /im userapi.exe    >nul 2>&1
taskkill /f /im userrpc.exe    >nul 2>&1
taskkill /f /im articleapi.exe >nul 2>&1
taskkill /f /im articlerpc.exe >nul 2>&1
taskkill /f /im identityapi.exe >nul 2>&1
taskkill /f /im identityrpc.exe >nul 2>&1

echo [2/3] Close Old CMD Windows...
taskkill /f /fi "WINDOWTITLE eq userapi"    >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq userrpc"    >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq articleapi" >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq articlerpc" >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq identityapi" >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq identityrpc" >nul 2>&1

rem 等 1 秒，确保 exe 文件锁和端口释放，否则 go build 覆盖 exe 会 Access is denied
timeout /t 1 /nobreak >nul

echo [3/3] go build .exe...
go build -o bin\userapi.exe    ./server/app/user/userapi       || goto :fail
go build -o bin\userrpc.exe    ./server/app/user/userrpc       || goto :fail
go build -o bin\articleapi.exe ./server/app/article/articleapi || goto :fail
go build -o bin\articlerpc.exe ./server/app/article/articlerpc || goto :fail
go build -o bin\identityapi.exe ./server/app/identity/identityapi || goto :fail
go build -o bin\identityrpc.exe ./server/app/identity/identityrpc || goto :fail

echo start...
start "userapi"    cmd /k bin\userapi.exe
start "userrpc"    cmd /k bin\userrpc.exe
start "articleapi" cmd /k bin\articleapi.exe
start "articlerpc" cmd /k bin\articlerpc.exe
start "identityapi" cmd /k bin\identityapi.exe
start "identityrpc" cmd /k bin\identityrpc.exe
goto :eof

:fail
echo.
echo go build err
pause
exit /b 1