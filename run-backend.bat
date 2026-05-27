@echo off
cd /d "%~dp0backend"
if not exist ..\.env (
  echo 请复制 ..\.env.example 为 ..\.env 并设置 DEEPSEEK_API_KEY
)
for /f "usebackq tokens=1,* delims==" %%a in ("..\.env") do (
  if not "%%a"=="" if not "%%a:~0,1%"=="#" set %%a=%%b
)
go run .
