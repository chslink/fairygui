# Godoc-MCP 自动配置脚本
# 此脚本将在 Claude Desktop 配置文件中添加 godoc-mcp

$ErrorActionPreference = "Stop"

# 获取用户配置路径
$ConfigPath = "$env:APPDATA\Claude\claude_desktop_config.json"
$BackupPath = "$env:APPDATA\Claude\claude_desktop_config.json.backup"

Write-Host "=== Godoc-MCP 自动配置脚本 ===" -ForegroundColor Green
Write-Host ""

# 检查 godoc-mcp 是否已安装
$GoBinPath = "$env:GOPATH\bin\godoc-mcp.exe"
if (!(Test-Path $GoBinPath)) {
    Write-Host "❌ 错误: 未找到 godoc-mcp，请先运行以下命令安装：" -ForegroundColor Red
    Write-Host "go install github.com/mrjoshuak/godoc-mcp@latest" -ForegroundColor Yellow
    exit 1
}
Write-Host "✓ 找到 godoc-mcp: $GoBinPath" -ForegroundColor Green

# 获取 Go 环境变量
$GoPath = $env:GOPATH
$GoModCache = $env:GOMODCACHE
if (!$GoPath) { $GoPath = "$env:USERPROFILE\go" }
if (!$GoModCache) { $GoModCache = "$env:USERPROFILE\go\pkg\mod" }

Write-Host "✓ Go Path: $GoPath" -ForegroundColor Green
Write-Host "✓ Go Mod Cache: $GoModCache" -ForegroundColor Green
Write-Host ""

# 创建配置内容
$ConfigJson = @{
    mcpServers = @{
        godoc = @{
            command = $GoBinPath.Replace('\', '\\')
            args = @()
            env = @{
                GOPATH = $GoPath.Replace('\', '\\')
                GOMODCACHE = $GoModCache.Replace('\', '\\')
            }
        }
    }
} | ConvertTo-Json -Depth 10

# 检查配置文件是否存在
if (Test-Path $ConfigPath) {
    Write-Host "检测到现有配置文件，正在备份..." -ForegroundColor Yellow
    Copy-Item $ConfigPath $BackupPath
    Write-Host "✓ 备份已保存到: $BackupPath" -ForegroundColor Green

    # 读取现有配置
    $ExistingConfig = Get-Content $ConfigPath | ConvertFrom-Json

    # 检查是否已存在 godoc 配置
    if ($ExistingConfig.PSObject.Properties.Name -contains "mcpServers" -and
        $ExistingConfig.mcpServers.PSObject.Properties.Name -contains "godoc") {
        Write-Host "⚠️  检测到 godoc 服务器已存在，将跳过添加" -ForegroundColor Yellow
        Write-Host "如需重新配置，请删除 godoc 条目后重试" -ForegroundColor Yellow
        exit 0
    }

    # 合并配置
    if ($ExistingConfig.PSObject.Properties.Name -notcontains "mcpServers") {
        $ExistingConfig | Add-Member -NotePropertyName "mcpServers" -NotePropertyValue @{}
    }

    foreach ($Property in $ConfigJson.PSObject.Properties) {
        $ExistingConfig.mcpServers | Add-Member -NotePropertyName $Property.Name -NotePropertyValue $Property.Value -Force
    }

    # 保存更新后的配置
    $UpdatedConfig = $ExistingConfig | ConvertTo-Json -Depth 10
    Set-Content -Path $ConfigPath -Value $UpdatedConfig -Encoding UTF8
} else {
    Write-Host "创建新配置文件..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Force -Path (Split-Path $ConfigPath)

    $ConfigJson | Set-Content -Path $ConfigPath -Encoding UTF8
}

Write-Host ""
Write-Host "✓ 配置完成！" -ForegroundColor Green
Write-Host ""
Write-Host "配置路径: $ConfigPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "接下来请：" -ForegroundColor Yellow
Write-Host "1. 重启 Claude Desktop 应用程序" -ForegroundColor White
Write-Host "2. 确认 MCP 服务器已连接" -ForegroundColor White
Write-Host ""
Write-Host "现在您可以这样问我：" -ForegroundColor Green
Write-Host '  "查看 pkg/fgui/core 包的文档"' -ForegroundColor Cyan
Write-Host '  "GObject 类型有哪些方法？"' -ForegroundColor Cyan
Write-Host ""
