param(
    [string]$PprofBaseUrl = "http://127.0.0.1:6060",
    [int]$CpuSeconds = 30,
    [string]$OutputDirectory = "loadtest/results",
    [string]$Prefix = "profile"
)

$ErrorActionPreference = "Stop"
$resolvedOutput = if ([IO.Path]::IsPathRooted($OutputDirectory)) {
    $OutputDirectory
} else {
    Join-Path $PSScriptRoot "..\$OutputDirectory"
}
New-Item -ItemType Directory -Force -Path $resolvedOutput | Out-Null

$safePrefix = $Prefix -replace '[^A-Za-z0-9._-]', '-'
$cpuPath = Join-Path $resolvedOutput "$safePrefix-cpu.pb.gz"
$heapPath = Join-Path $resolvedOutput "$safePrefix-heap.pb.gz"
$goroutinePath = Join-Path $resolvedOutput "$safePrefix-goroutines.pprof.txt"
$cpuTopPath = Join-Path $resolvedOutput "$safePrefix-cpu-top.txt"
$heapTopPath = Join-Path $resolvedOutput "$safePrefix-heap-top.txt"

$pprof = Get-Command pprof -ErrorAction SilentlyContinue
if (-not $pprof) {
    $workspacePprof = Join-Path $PSScriptRoot "..\..\.tools\pprof.exe"
    if (Test-Path $workspacePprof) {
        $pprof = Get-Item $workspacePprof
    }
}
if (-not $pprof) {
    throw "pprof CLI not found; install with: go install github.com/google/pprof@latest"
}
$pprofPath = if ($pprof.Source) { $pprof.Source } else { $pprof.FullName }

function Invoke-PprofCommand {
    param([string[]]$Arguments)
    $previousErrorActionPreference = $ErrorActionPreference
    try {
        # pprof writes normal progress messages to stderr. PowerShell 5 wraps
        # them as NativeCommandError records even when the process exits 0.
        $ErrorActionPreference = "Continue"
        $commandOutput = & $pprofPath @Arguments 2>&1
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }
    if ($exitCode -ne 0) {
        throw "pprof failed with exit code $exitCode`: $($commandOutput -join [Environment]::NewLine)"
    }
}

function Remove-WorkspacePath {
    param([string]$Path)
    $workspaceRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
    $content = [IO.File]::ReadAllText($Path)
    $content = $content.Replace($workspaceRoot, "<workspace>")
    [IO.File]::WriteAllText($Path, $content, [Text.UTF8Encoding]::new($false))
}

$previousPprofTmpDir = $env:PPROF_TMPDIR
$env:PPROF_TMPDIR = $resolvedOutput
try {
    Invoke-PprofCommand @("-proto", "-seconds", [string]$CpuSeconds, "-output", $cpuPath, "$PprofBaseUrl/debug/pprof/profile")
    Invoke-PprofCommand @("-proto", "-output", $heapPath, "$PprofBaseUrl/debug/pprof/heap")
    Invoke-PprofCommand @("-top", "-nodecount", "20", "-output", $cpuTopPath, $cpuPath)
    Invoke-PprofCommand @("-top", "-sample_index", "inuse_space", "-nodecount", "20", "-output", $heapTopPath, $heapPath)
    Remove-WorkspacePath $cpuTopPath
    Remove-WorkspacePath $heapTopPath
} finally {
    $env:PPROF_TMPDIR = $previousPprofTmpDir
}
Invoke-WebRequest -UseBasicParsing "$PprofBaseUrl/debug/pprof/goroutine?debug=1" -OutFile $goroutinePath

Write-Host "Profiles written to $resolvedOutput"
Write-Host "Inspect CPU:  pprof -http=:0 $cpuPath"
Write-Host "Inspect heap: pprof -http=:0 $heapPath"
Write-Host "Versionable reports: $cpuTopPath and $heapTopPath"
