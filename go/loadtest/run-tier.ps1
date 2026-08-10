param(
    [Parameter(Mandatory = $true)]
    [ValidateSet(10, 100, 500)]
    [int]$Listeners,
    [Parameter(Mandatory = $true)]
    [string]$BaseUrl,
    [Parameter(Mandatory = $true)]
    [string]$StreamId,
    [string]$ListenerToken = "",
    [string]$MetricsBearerToken = "",
    [string]$PprofBaseUrl = "",
    [int]$CpuProfileSeconds = 30,
    [int]$RestSeconds = 35,
    [int]$SampleIntervalSeconds = 2,
    [string]$OutputDirectory = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

if (-not $OutputDirectory) {
    $OutputDirectory = Join-Path $PSScriptRoot "results"
}
$OutputDirectory = [IO.Path]::GetFullPath($OutputDirectory)
New-Item -ItemType Directory -Force -Path $OutputDirectory | Out-Null

$k6 = Get-Command k6 -ErrorAction SilentlyContinue
if (-not $k6) {
    throw "k6 is not installed or is not available in PATH"
}

$base = [Uri]$BaseUrl
if ($base.Scheme -notin @("http", "https")) {
    throw "BaseUrl must use HTTP or HTTPS"
}
$baseUrlNormalized = $BaseUrl.TrimEnd("/")
$metricsUrl = "$($base.Scheme)://$($base.Authority)/metrics"
$tierPrefix = "tier-$Listeners"
$samplesPath = Join-Path $OutputDirectory "$tierPrefix-runtime.csv"
$stdoutPath = Join-Path $OutputDirectory "$tierPrefix-k6.log"
$stderrPath = Join-Path $OutputDirectory "$tierPrefix-k6-errors.log"
$summaryPath = Join-Path $OutputDirectory "k6-$Listeners-summary.json"

function Get-MetricValue {
    param([string]$Text, [string]$Name)
    $match = [regex]::Match(
        $Text,
        "(?m)^$([regex]::Escape($Name))(?:\{[^}]*\})?\s+([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)`r?$"
    )
    if (-not $match.Success) { return $null }
    return [double]::Parse($match.Groups[1].Value, [Globalization.CultureInfo]::InvariantCulture)
}

function Get-MetricSum {
    param(
        [string]$Text,
        [string]$Name,
        [string]$StreamId,
        [string]$RequiredLabel = ""
    )
    $sum = 0.0
    $found = $false
    foreach ($line in ($Text -split "`n")) {
        if ($line -notmatch "^$([regex]::Escape($Name))\{([^}]*)\}\s+([-+]?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?)`r?$") {
            continue
        }
        $labels = $matches[1]
        $value = $matches[2]
        if ($labels -notmatch "stream_id=`"$([regex]::Escape($StreamId))`"") {
            continue
        }
        if ($RequiredLabel -and $labels -notmatch [regex]::Escape($RequiredLabel)) {
            continue
        }
        $sum += [double]::Parse($value, [Globalization.CultureInfo]::InvariantCulture)
        $found = $true
    }
    if (-not $found) { return $null }
    return $sum
}

function Read-RuntimeSample {
    $headers = @{}
    if ($MetricsBearerToken) {
        $headers.Authorization = "Bearer $MetricsBearerToken"
    }
    $response = Invoke-WebRequest -UseBasicParsing -Uri $metricsUrl -Headers $headers -TimeoutSec 15
    $text = [string]$response.Content
    [pscustomobject]@{
        TimestampUtc = [DateTimeOffset]::UtcNow.ToString("o")
        EpochSeconds = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds() / 1000.0
        CpuSeconds = Get-MetricValue $text "process_cpu_seconds_total"
        RssBytes = Get-MetricValue $text "process_resident_memory_bytes"
        Goroutines = Get-MetricValue $text "go_goroutines"
        DroppedChunks = Get-MetricSum $text "streampulse_audio_dropped_chunks_total" $StreamId
        IngestChunks = Get-MetricSum $text "streampulse_audio_chunks_total" $StreamId 'direction="ingest"'
    }
}

function Get-Percentile95 {
    param([double[]]$Values)
    if (-not $Values -or $Values.Count -eq 0) { return $null }
    $ordered = @($Values | Sort-Object)
    $index = [Math]::Max(0, [Math]::Ceiling(0.95 * $ordered.Count) - 1)
    return $ordered[$index]
}

$environmentKeys = @(
    "LISTENERS", "BASE_URL", "STREAM_ID", "LISTENER_TOKEN", "LISTENER_PATH",
    "RESULTS_DIR", "RESULT_PREFIX", "START_RAMP_SECONDS"
)
$previousEnvironment = @{}
foreach ($key in $environmentKeys) {
    $previousEnvironment[$key] = [Environment]::GetEnvironmentVariable($key, "Process")
}

$goDirectory = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot ".."))
$goDirectoryUri = [Uri]($goDirectory.TrimEnd("\") + "\")
$outputDirectoryUri = [Uri]($OutputDirectory.TrimEnd("\") + "\")
$relativeResults = $goDirectoryUri.MakeRelativeUri($outputDirectoryUri).ToString().TrimEnd("/")
[Environment]::SetEnvironmentVariable("LISTENERS", [string]$Listeners, "Process")
[Environment]::SetEnvironmentVariable("BASE_URL", $baseUrlNormalized, "Process")
[Environment]::SetEnvironmentVariable("STREAM_ID", $StreamId, "Process")
[Environment]::SetEnvironmentVariable("LISTENER_TOKEN", $ListenerToken, "Process")
[Environment]::SetEnvironmentVariable("LISTENER_PATH", $(if ($ListenerToken) { "listen" } else { "audio" }), "Process")
[Environment]::SetEnvironmentVariable("RESULTS_DIR", $relativeResults, "Process")
[Environment]::SetEnvironmentVariable("RESULT_PREFIX", "k6-$Listeners", "Process")
[Environment]::SetEnvironmentVariable("START_RAMP_SECONDS", "10", "Process")

$profileJob = $null
$k6Process = $null
$k6Started = $false
$samples = [Collections.Generic.List[object]]::new()
try {
    # Start-Process can expose a null ExitCode under constrained Windows
    # sessions. A direct Process object preserves the real k6 exit status while
    # asynchronous reads prevent its stdout/stderr pipes from blocking.
    $processInfo = [Diagnostics.ProcessStartInfo]::new()
    $processInfo.FileName = $k6.Source
    $processInfo.Arguments = "run loadtest/stream.js"
    $processInfo.WorkingDirectory = $goDirectory
    $processInfo.UseShellExecute = $false
    $processInfo.CreateNoWindow = $true
    $processInfo.RedirectStandardOutput = $true
    $processInfo.RedirectStandardError = $true
    $processInfo.StandardOutputEncoding = [Text.UTF8Encoding]::new($false)
    $processInfo.StandardErrorEncoding = [Text.UTF8Encoding]::new($false)
    $k6Process = [Diagnostics.Process]::new()
    $k6Process.StartInfo = $processInfo
    if (-not $k6Process.Start()) {
        throw "failed to start k6"
    }
    $k6Started = $true
    $stdoutTask = $k6Process.StandardOutput.ReadToEndAsync()
    $stderrTask = $k6Process.StandardError.ReadToEndAsync()

    if ($PprofBaseUrl) {
        $captureScript = Join-Path $PSScriptRoot "capture-pprof.ps1"
        $profileJob = Start-Job -ScriptBlock {
            param($Script, $Url, $Directory, $Seconds, $Prefix)
            & $Script -PprofBaseUrl $Url -OutputDirectory $Directory -CpuSeconds $Seconds -Prefix $Prefix
        } -ArgumentList $captureScript, $PprofBaseUrl, $OutputDirectory, $CpuProfileSeconds, $tierPrefix
    }

    while (-not $k6Process.HasExited) {
        try {
            $samples.Add((Read-RuntimeSample))
        } catch {
            Write-Warning "Runtime metrics sample failed: $($_.Exception.Message)"
        }
        Start-Sleep -Seconds $SampleIntervalSeconds
        $k6Process.Refresh()
    }
    $k6Process.WaitForExit()
    $k6ExitCode = [int]$k6Process.ExitCode
    [IO.File]::WriteAllText($stdoutPath, [string]$stdoutTask.Result, [Text.UTF8Encoding]::new($false))
    [IO.File]::WriteAllText($stderrPath, [string]$stderrTask.Result, [Text.UTF8Encoding]::new($false))
    if (-not (Test-Path $summaryPath)) {
        throw "k6 did not produce $summaryPath"
    }

    if ($profileJob) {
        Wait-Job $profileJob | Out-Null
        Receive-Job $profileJob
        if ($profileJob.State -ne "Completed") {
            throw "pprof capture failed with state $($profileJob.State)"
        }
    }

    Start-Sleep -Seconds $RestSeconds
    try {
        $samples.Add((Read-RuntimeSample))
    } catch {
        Write-Warning "Post-rest metrics sample failed: $($_.Exception.Message)"
    }
} finally {
    if ($k6Started -and -not $k6Process.HasExited) {
        $k6Process.Kill()
        $k6Process.WaitForExit()
    }
    if ($k6Process) {
        $k6Process.Dispose()
    }
    if ($profileJob) {
        Remove-Job $profileJob -Force -ErrorAction SilentlyContinue
    }
    foreach ($key in $environmentKeys) {
        [Environment]::SetEnvironmentVariable($key, $previousEnvironment[$key], "Process")
    }
}

if ($samples.Count -lt 2) {
    throw "At least two runtime metric samples are required"
}
$samples | Export-Csv -NoTypeInformation -Encoding UTF8 -Path $samplesPath

$cpuPercentages = [Collections.Generic.List[double]]::new()
for ($index = 1; $index -lt $samples.Count; $index++) {
    $previous = $samples[$index - 1]
    $current = $samples[$index]
    if ($null -eq $previous.CpuSeconds -or $null -eq $current.CpuSeconds) { continue }
    $elapsed = $current.EpochSeconds - $previous.EpochSeconds
    if ($elapsed -le 0) { continue }
    $cpuPercentages.Add((($current.CpuSeconds - $previous.CpuSeconds) / $elapsed) * 100.0)
}
$cpuArray = $cpuPercentages.ToArray()

$k6Summary = Get-Content -Raw $summaryPath | ConvertFrom-Json
$failedThresholds = @(
    $k6Summary.metrics.psobject.Properties |
        ForEach-Object {
            $thresholdProperty = $_.Value.psobject.Properties["thresholds"]
            if ($null -ne $thresholdProperty) {
                $thresholdProperty.Value.psobject.Properties
            }
        } |
        Where-Object { $_.Value.ok -eq $false } |
        ForEach-Object { $_.Name }
)
$first = $samples[0]
$last = $samples[$samples.Count - 1]
$dropDelta = if ($null -ne $first.DroppedChunks -and $null -ne $last.DroppedChunks) {
    [Math]::Max(0, $last.DroppedChunks - $first.DroppedChunks)
} else { $null }
$ingestDelta = if ($null -ne $first.IngestChunks -and $null -ne $last.IngestChunks) {
    [Math]::Max(0, $last.IngestChunks - $first.IngestChunks)
} else { $null }
if ($null -eq $dropDelta -and $ingestDelta -gt 0) {
    # A Prometheus counter vector has no time series until its first increment.
    # Ingest with no dropped series is therefore an observed zero, not missing data.
    $dropDelta = 0.0
}
$dropRate = if ($null -ne $dropDelta -and $ingestDelta -gt 0) { 100.0 * $dropDelta / $ingestDelta } else { $null }
$rssValues = @($samples | Where-Object { $null -ne $_.RssBytes } | ForEach-Object { [double]$_.RssBytes })
$publisherSummaryPath = Join-Path $OutputDirectory "publisher-$Listeners-summary.json"
$sourcePayloadKbit = $null
if (Test-Path $publisherSummaryPath) {
    $publisherSummary = Get-Content -Raw $publisherSummaryPath | ConvertFrom-Json
    $publisherMetadata = $publisherSummary.psobject.Properties["publisher"]
    if ($null -ne $publisherMetadata) {
        $sourcePayloadKbit = [double]$publisherMetadata.Value.measured_payload_kbit_s
    }
}

$summary = [ordered]@{
    timestamp_utc = [DateTimeOffset]::UtcNow.ToString("o")
    target = "$($base.Scheme)://$($base.Authority)"
    listeners = $Listeners
    listener_path = $(if ($ListenerToken) { "listen" } else { "audio" })
    source_payload_kbit_s = $sourcePayloadKbit
    cpu_p95_percent_of_one_core = Get-Percentile95 $cpuArray
    rss_max_bytes = $(if ($rssValues.Count) { ($rssValues | Measure-Object -Maximum).Maximum } else { $null })
    dropped_chunks = $dropDelta
    dropped_chunks_percent_of_ingest = $dropRate
    k6_checks_percent = 100.0 * [double]$k6Summary.metrics.checks.values.rate
    k6_failed_requests_percent = 100.0 * [double]$k6Summary.metrics.http_req_failed.values.rate
    k6_http_duration_p95_ms = [double]$k6Summary.metrics.http_req_duration.values.'p(95)'
    k6_exit_code = $k6ExitCode
    k6_thresholds_ok = ($failedThresholds.Count -eq 0)
    goroutines_after_rest = $last.Goroutines
    rest_seconds = $RestSeconds
}
$tierSummaryPath = Join-Path $OutputDirectory "$tierPrefix-summary.json"
$summary | ConvertTo-Json -Depth 5 | Set-Content -Encoding UTF8 $tierSummaryPath

$tierFiles = Get-ChildItem $OutputDirectory -Filter "tier-*-summary.json" |
    Sort-Object { [int](($_.BaseName -replace '^tier-', '') -replace '-summary$', '') }
$report = @(
    "# k6 + runtime evidence",
    "",
    "- Target: $($base.Scheme)://$($base.Authority)",
    "- Scope: runtime metrics and pprof belong to the explicitly configured target; localhost is not Render production.",
    "",
    "| Palier | Source (kbit/s) | CPU p95 (% d'un coeur) | RSS max (MiB) | Drops | k6 checks | Requetes echouees | Duree HTTP p95 | Goroutines apres repos | Statut |",
    "| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | :---: |"
)
foreach ($file in $tierFiles) {
    $row = Get-Content -Raw $file.FullName | ConvertFrom-Json
    $rssMiB = if ($null -ne $row.rss_max_bytes) { [Math]::Round($row.rss_max_bytes / 1MB, 2) } else { "n/a" }
    $cpu = if ($null -ne $row.cpu_p95_percent_of_one_core) { [Math]::Round($row.cpu_p95_percent_of_one_core, 2) } else { "n/a" }
    $drops = if ($null -ne $row.dropped_chunks_percent_of_ingest) { "$([Math]::Round($row.dropped_chunks_percent_of_ingest, 4)) %" } else { "n/a" }
    $checks = "$([Math]::Round($row.k6_checks_percent, 3)) %"
    $failed = "$([Math]::Round($row.k6_failed_requests_percent, 3)) %"
    $duration = "$([Math]::Round($row.k6_http_duration_p95_ms / 1000.0, 3)) s"
    $source = if ($null -ne $row.source_payload_kbit_s) { [Math]::Round($row.source_payload_kbit_s, 2) } else { "n/a" }
    $status = if ($row.k6_thresholds_ok) { "OK" } else { "ECHEC" }
    $report += "| $($row.listeners) | $source | $cpu | $rssMiB | $drops | $checks | $failed | $duration | $($row.goroutines_after_rest) | $status |"
}
$report | Set-Content -Encoding UTF8 (Join-Path $OutputDirectory "summary.md")

$summary | Format-List
if (($null -ne $k6ExitCode -and $k6ExitCode -ne 0) -or $failedThresholds.Count -gt 0) {
    throw "k6 failed (exit=$k6ExitCode, thresholds=$($failedThresholds -join ',')); inspect $stderrPath"
}
