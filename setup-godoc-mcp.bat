@echo off
chcp 65001 >nul
echo === Godoc-MCP 自动配置脚本 ===
echo.

REM 检查 PowerShell
where powershell >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未找到 PowerShell
    echo 请手动添加 MCP 配置
    pause
    exit /b 1
)

REM 运行 PowerShell 脚本
powershell -ExecutionPolicy Bypass -File "%~dp0setup-godoc-mcp.ps1"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo 脚本执行失败，请检查错误信息
    pause
    exit /b 1
)

echo.
echo 配置完成！按任意键退出...
pause >nul
