# Watchtower 安全扫描脚本 (Windows PowerShell)
# 用于检查依赖中的安全漏洞和过期依赖

Write-Host "🔍 开始 Watchtower 安全扫描..." -ForegroundColor Cyan
Write-Host ""

# 1. 检查 Go 版本
Write-Host "📋 检查 Go 版本..." -ForegroundColor Blue
go version
Write-Host ""

# 2. 检查依赖漏洞
Write-Host "🔒 检查依赖漏洞 (govulncheck)..." -ForegroundColor Blue
$govulncheck = Join-Path $env:USERPROFILE "go\bin\govulncheck.exe"
if (-not (Test-Path $govulncheck)) {
    Write-Host "安装 govulncheck..." -ForegroundColor Yellow
    go install golang.org/x/vuln/cmd/govulncheck@latest
}
& $govulncheck ./...
Write-Host ""

# 3. 检查代码安全问题
Write-Host "🛡️ 检查代码安全问题 (gosec)..." -ForegroundColor Blue
$gosec = Join-Path $env:USERPROFILE "go\bin\gosec.exe"
if (-not (Test-Path $gosec)) {
    Write-Host "安装 gosec..." -ForegroundColor Yellow
    go install github.com/securego/gosec/v2/cmd/gosec@latest
}
& $gosec ./... 2>&1 | Select-String -Pattern "No issues found" -Context 0,10
Write-Host ""

# 4. 检查过期依赖
Write-Host "📦 检查过期依赖..." -ForegroundColor Blue
$goModOutdated = Join-Path $env:USERPROFILE "go\bin\go-mod-outdated.exe"
if (-not (Test-Path $goModOutdated)) {
    Write-Host "安装 go-mod-outdated..." -ForegroundColor Yellow
    go install github.com/psampaz/go-mod-outdated@latest
}
go list -u -m -json all | & $goModOutdated -update -direct
Write-Host ""

# 5. 依赖统计
Write-Host "📊 依赖统计" -ForegroundColor Blue
$totalDeps = (go list -m all | Measure-Object).Count
$directDeps = (go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all | Measure-Object).Count
Write-Host "总依赖数: $totalDeps"
Write-Host "直接依赖数: $directDeps"
Write-Host ""

# 6. 检查许可证
Write-Host "📄 检查许可证..." -ForegroundColor Blue
$goLicenses = Join-Path $env:USERPROFILE "go\bin\go-licenses.exe"
if (-not (Test-Path $goLicenses)) {
    Write-Host "安装 go-licenses..." -ForegroundColor Yellow
    go install github.com/google/go-licenses@latest
}
Write-Host "生成许可证报告..."
& $goLicenses save ./... --save_path=./licenses 2>&1 | Out-Null
Write-Host ""

Write-Host "✅ 安全扫描完成！" -ForegroundColor Green
Write-Host ""
Write-Host "📝 建议：" -ForegroundColor Cyan
Write-Host "1. 定期运行此脚本检查安全问题"
Write-Host "2. 及时更新发现漏洞的依赖"
Write-Host "3. 关注 GitHub Actions 的安全扫描结果"
Write-Host "4. 定期审查第三方依赖的许可证合规性"