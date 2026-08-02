param(
    [string]$PprofBaseUrl = "http://127.0.0.1:6060",
    [int]$CpuSeconds = 30,
    [string]$OutputDirectory = "loadtest/results"
)

$ErrorActionPreference = "Stop"
$resolvedOutput = Join-Path $PSScriptRoot "..\$OutputDirectory"
New-Item -ItemType Directory -Force -Path $resolvedOutput | Out-Null

$cpuPath = Join-Path $resolvedOutput "cpu.pb.gz"
$heapPath = Join-Path $resolvedOutput "heap.pb.gz"
$goroutinePath = Join-Path $resolvedOutput "goroutine.txt"

go tool pprof -proto -seconds $CpuSeconds -output $cpuPath "$PprofBaseUrl/debug/pprof/profile"
go tool pprof -proto -output $heapPath "$PprofBaseUrl/debug/pprof/heap"
Invoke-WebRequest -UseBasicParsing "$PprofBaseUrl/debug/pprof/goroutine?debug=1" -OutFile $goroutinePath

Write-Host "Profiles written to $resolvedOutput"
Write-Host "Inspect CPU:  go tool pprof -http=:0 $cpuPath"
Write-Host "Inspect heap: go tool pprof -http=:0 $heapPath"
