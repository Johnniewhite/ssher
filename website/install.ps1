$ErrorActionPreference = "Stop"

$architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
switch ($architecture) {
    "x64"   { $releaseArchitecture = "amd64" }
    "arm64" { $releaseArchitecture = "arm64" }
    default  { throw "ssher does not currently provide a Windows build for $architecture." }
}

$asset = "ssher_Windows_$releaseArchitecture.zip"
$releaseBase = "https://github.com/Johnniewhite/ssher/releases/latest/download"
$installDirectory = Join-Path $env:LOCALAPPDATA "Programs\ssher"
$temporaryDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ("ssher-install-" + [guid]::NewGuid())
$archive = Join-Path $temporaryDirectory $asset
$checksums = Join-Path $temporaryDirectory "checksums.txt"

try {
    New-Item -ItemType Directory -Path $temporaryDirectory -Force | Out-Null
    Invoke-WebRequest "$releaseBase/$asset" -OutFile $archive -UseBasicParsing
    Invoke-WebRequest "$releaseBase/checksums.txt" -OutFile $checksums -UseBasicParsing

    $checksumLine = Get-Content $checksums | Where-Object { $_ -match "\s+$([regex]::Escape($asset))$" } | Select-Object -First 1
    if (-not $checksumLine) {
        throw "The release checksum for $asset was not found."
    }
    $expected = ($checksumLine -split "\s+")[0].ToLowerInvariant()
    $actual = (Get-FileHash -Path $archive -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        throw "Checksum verification failed for $asset."
    }

    New-Item -ItemType Directory -Path $installDirectory -Force | Out-Null
    Expand-Archive -Path $archive -DestinationPath $installDirectory -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $pathEntries = @($userPath -split ";" | Where-Object { $_ })
    if ($pathEntries -notcontains $installDirectory) {
        $updatedPath = (@($pathEntries) + $installDirectory) -join ";"
        [Environment]::SetEnvironmentVariable("Path", $updatedPath, "User")
    }
    if (($env:Path -split ";") -notcontains $installDirectory) {
        $env:Path = "$env:Path;$installDirectory"
    }

    & (Join-Path $installDirectory "ssher.exe") --version
    Write-Host "ssher is installed. Open a new PowerShell window and run: ssher" -ForegroundColor Green
}
finally {
    if (Test-Path $temporaryDirectory) {
        Remove-Item -Path $temporaryDirectory -Recurse -Force
    }
}
