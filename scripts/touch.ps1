param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]] $Paths
)

foreach ($path in $Paths) {
    try {
        if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
            New-Item -ItemType File -Path $path -Force | Out-Null
        } else {
            $now = Get-Date
            [System.IO.File]::SetLastAccessTime($path, $now)
            [System.IO.File]::SetLastWriteTime($path, $now)
        }
        Write-Output "touched $path"
    }
    catch {
        Write-Error "failed to touch '$path': $($_.Exception.Message)"
        exit 1
    }
}
