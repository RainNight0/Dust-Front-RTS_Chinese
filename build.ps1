# Dust Front RTS Demo Chinese Localization Patch Build Script
# Usage:  powershell -ExecutionPolicy Bypass -File build.ps1
#         powershell -ExecutionPolicy Bypass -File build.ps1 -OutFile "dist\patch.exe"
[CmdletBinding()]
param(
    # 输出 exe 的路径；相对路径以本脚本所在目录为基准
    [string]$OutFile = "Dust Front RTS Demo汉化补丁.exe"
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

# 只在外部未指定时才套用国内镜像，方便 CI 覆盖成官方源
if (-not $env:GOPROXY) { $env:GOPROXY = "https://goproxy.cn,direct" }
if (-not $env:GOSUMDB) { $env:GOSUMDB = "off" }

# walk needs a manifest with comctl32 v6 (otherwise "TTM_ADDTOOL failed" at startup)
$goBinDir = if ($env:GOBIN) { $env:GOBIN } else { Join-Path (go env GOPATH) "bin" }
$rsrc = Join-Path $goBinDir "rsrc.exe"
if (-not (Test-Path $rsrc)) {
    Write-Host "installing rsrc..."
    go install github.com/akavel/rsrc@latest
    if ($LASTEXITCODE -ne 0) { throw "installing rsrc failed" }
}
& $rsrc -manifest app.manifest -arch amd64 -o rsrc.syso
if ($LASTEXITCODE -ne 0) { throw "rsrc failed" }

# NOTE: the Chinese name below must survive this .ps1 file's encoding.
# If it gets mangled, run the go build line manually instead.
$outDir = Split-Path -Parent $OutFile
if ($outDir) { New-Item -ItemType Directory -Path $outDir -Force | Out-Null }
go build -ldflags "-H windowsgui -s -w" -o $OutFile .
if ($LASTEXITCODE -ne 0) { throw "build failed" }

Write-Host ""
Write-Host "BUILD OK -> $PSScriptRoot\$OutFile"
Write-Host "将该exe文件复制到“Dust Front RTS Demo”文件夹中，然后运行它。"
