$ErrorActionPreference = 'Stop'

$repo = "leolebleis/scpclip"
$installDir = if ($env:SCPCLIP_INSTALL_DIR) { $env:SCPCLIP_INSTALL_DIR } else { Join-Path $HOME ".local\bin" }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$versionTrimmed = $version.TrimStart('v')

$archive = "scpclip_${versionTrimmed}_windows_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$version/$archive"
$checksumsUrl = "https://github.com/$repo/releases/download/$version/checksums.txt"

$tmpDir = Join-Path ([IO.Path]::GetTempPath()) "scpclip-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    Write-Host "Installing scpclip $version (windows/$arch)..."
    Invoke-WebRequest $url -OutFile (Join-Path $tmpDir $archive) -UseBasicParsing
    Invoke-WebRequest $checksumsUrl -OutFile (Join-Path $tmpDir "checksums.txt") -UseBasicParsing

    $expected = ((Get-Content (Join-Path $tmpDir "checksums.txt")) -match $archive) -replace '\s+.*',''
    $actual = (Get-FileHash (Join-Path $tmpDir $archive) -Algorithm SHA256).Hash.ToLower()
    if ($expected -ne $actual) {
        throw "Checksum mismatch: expected $expected, got $actual"
    }

    Expand-Archive (Join-Path $tmpDir $archive) -DestinationPath $tmpDir -Force
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Move-Item (Join-Path $tmpDir "scpclip.exe") (Join-Path $installDir "scpclip.exe") -Force

    Write-Host "Installed scpclip $version to $installDir\scpclip.exe"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$installDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
        Write-Host "Added $installDir to user PATH (restart your terminal)."
    }
} finally {
    Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
