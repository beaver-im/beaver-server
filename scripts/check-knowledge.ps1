# Documentation-only validator. Uses PowerShell 5.1+ and .NET; no installs/network/codegen.
[CmdletBinding()]
param([switch]$BeforeLegacyRemoval)
$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$utf8 = New-Object Text.UTF8Encoding($false)
$manifest = [IO.File]::ReadAllText((Join-Path $root 'docs/knowledge-checks.json'),$utf8) | ConvertFrom-Json
$errors = New-Object 'Collections.Generic.List[string]'
$warnings = New-Object 'Collections.Generic.List[string]'
$counts = @{links=0;paths=0;commands=0;sources=0;documents=0}
function Fail([string]$message) { $errors.Add($message) }
function Resolve-Local([string]$base,[string]$relative) {
    [IO.Path]::GetFullPath((Join-Path $base ([Uri]::UnescapeDataString($relative))))
}
function Strip-Fences([string]$text) {
    $lines = $text -split "`r?`n"
    $inside = $false; $fence = ''; $length=0
    $result = New-Object 'Collections.Generic.List[string]'
    foreach ($line in $lines) {
        if ($line -match '^\s*(`{3,}|~{3,})') {
            $mark=$matches[1]
            if (-not $inside) {$inside=$true;$fence=$mark.Substring(0,1);$length=$mark.Length}
            elseif ($mark.Substring(0,1) -eq $fence -and $mark.Length -ge $length) {$inside=$false}
            continue
        }
        if (-not $inside) {$result.Add($line)}
    }
    if ($inside) {Fail 'Unclosed Markdown code fence'}
    $result -join "`n"
}
foreach ($relative in $manifest.documents) {
    $file=Resolve-Local $root $relative
    if (-not (Test-Path -LiteralPath $file -PathType Leaf)) {Fail "Missing document: $relative";continue}
    $counts.documents++
    $text=[IO.File]::ReadAllText($file,$utf8)
    if($relative -like 'docs/appendices/source-material/*'){$text=Strip-Fences $text}
    if ($text -match '(?m)^(<<<<<<< |=======\s*$|>>>>>>> )') {Fail "Conflict marker: $relative"}
    if ($text -match '(?m)[\t ]+\r?$') {Fail "Trailing whitespace: $relative"}
    if ($text -match '\]\([^\r\n]*\\') {Fail "Use slash in Markdown destination: $relative"}
    $active=Strip-Fences $text
    foreach ($m in [regex]::Matches($active,'(?<!!)\[[^\]\r\n]+\]\(([^\s)]+)\)')) {
        $target=$m.Groups[1].Value
        if ($target -match '^(https?://|mailto:|sand-workflow:)') {continue}
        $parts=$target -split '#',2
        $path=$parts[0]
        $dest=if($path){Resolve-Local (Split-Path $file -Parent) $path}else{$file}
        $counts.links++
        $isSibling=$dest -like (([IO.Path]::GetFullPath((Join-Path $root '../beaver-server-package')))+'*')
        if (-not(Test-Path -LiteralPath $dest)) {
            if($isSibling){$warnings.Add("Optional sibling unavailable: $target")}else{Fail "Broken link in ${relative}: $target"}
            continue
        }
        if ($parts.Count -eq 2 -and $parts[1]) {
            $fragment=[Uri]::UnescapeDataString($parts[1])
            $destText=[IO.File]::ReadAllText($dest,$utf8)
            $anchors=@([regex]::Matches($destText,'(?m)^#{1,6}\s+(.+)$') | ForEach-Object {
                (($_.Groups[1].Value.ToLowerInvariant() -replace '[`*_]','') -replace '[^\p{L}\p{N}\s_-]','') -replace '\s','-'
            })
            if ($fragment -notin $anchors -and $destText -notmatch ('id=["'']'+[regex]::Escape($fragment)+'["'']')) {Fail "Missing heading in ${relative}: $target"}
        }
    }
    if($relative -notlike 'docs/appendices/*' -and $active -match 'skills/beaver-dev') {Fail "Stale active legacy path: $relative"}
}
foreach($path in $manifest.paths){
    $counts.paths++
    if(-not(Test-Path -LiteralPath (Resolve-Local $root $path))){Fail "Missing registered source path: $path"}
}
foreach($c in $manifest.commands){
    $counts.commands++
    if(-not(Test-Path -LiteralPath (Resolve-Local $root $c.cwd) -PathType Container)){Fail "Missing command cwd: $($c.id)"}
    foreach($p in $c.requiredPaths){if(-not(Test-Path -LiteralPath (Resolve-Local (Resolve-Local $root $c.cwd) $p))){Fail "Missing command input: $($c.id) / $p"}}
    foreach($e in $c.evidence){
        $f=Resolve-Local $root $e.file
        if(-not(Test-Path -LiteralPath $f)){Fail "Missing command evidence: $($c.id)";continue}
        if(-not([IO.File]::ReadAllText($f,$utf8).Contains($e.contains))){Fail "Changed command evidence: $($c.id) / $($e.file)"}
    }
    if(-not(Get-Command $c.executable -ErrorAction SilentlyContinue)){$warnings.Add("Tool not on PATH (not installed by validator): $($c.executable)")}
}
$begin="<!-- source-text-begin -->`n````````text`n"
$end="`n`````````n<!-- source-text-end -->"
foreach($s in $manifest.migration){
    $counts.sources++
    $file=Resolve-Local $root $s.snapshot
    if(-not(Test-Path -LiteralPath $file)){Fail "Missing frozen source: $($s.source)";continue}
    $text=[IO.File]::ReadAllText($file,$utf8)
    $a=$text.IndexOf($begin,[StringComparison]::Ordinal)
    $b=$text.IndexOf($end,[StringComparison]::Ordinal)
    if($a -lt 0 -or $b -lt $a){Fail "Bad frozen-source delimiters: $($s.snapshot)";continue}
    $raw=$text.Substring($a+$begin.Length,$b-($a+$begin.Length))
    $sha=[Security.Cryptography.SHA256]::Create()
    try{$hash=([BitConverter]::ToString($sha.ComputeHash($utf8.GetBytes($raw)))).Replace('-','')}finally{$sha.Dispose()}
    if($hash -ne $s.sha256){Fail "Frozen source hash mismatch: $($s.source)"}
    $old=Resolve-Local $root $s.source
    if($BeforeLegacyRemoval){
        if(-not(Test-Path -LiteralPath $old)){Fail "Source disappeared before removal: $($s.source)"}
        elseif((Get-FileHash -LiteralPath $old -Algorithm SHA256).Hash -ne $s.sha256){Fail "Concurrent legacy edit: $($s.source)"}
    }elseif(Test-Path -LiteralPath $old){Fail "Legacy source still present: $($s.source)"}
}
if(@($manifest.migration).Count -ne 7){Fail 'Expected all seven legacy source files'}
$owned=@($manifest.documents)+@('scripts/check-knowledge.ps1','docs/knowledge-checks.json')
foreach($relative in $owned){
    $text=[IO.File]::ReadAllText((Resolve-Local $root $relative),$utf8)
    if($relative -like 'docs/appendices/source-material/*'){$text=Strip-Fences $text}
    if($text -match '(?m)^(<<<<<<< |=======\s*$|>>>>>>> )'){Fail "Conflict marker: $relative"}
    if($text -match '(?m)[\t ]+\r?$'){Fail "Trailing whitespace: $relative"}
}
$ledger=[IO.File]::ReadAllText((Join-Path $root 'docs/appendices/migration.md'),$utf8)
foreach($id in $manifest.migrationTopics){
    if($ledger -notmatch ('\| '+[regex]::Escape($id)+' ')){Fail "Missing migration topic: $id"}
}
foreach($dir in $manifest.managedDocDirectories){
    foreach($f in Get-ChildItem -LiteralPath (Resolve-Local $root $dir) -Recurse -File -Filter '*.md'){
        $relative=$f.FullName.Substring($root.Length+1).Replace('\','/')
        if($relative -notin $manifest.documents){Fail "Register new handbook document: $relative"}
    }
}
foreach($w in $warnings){Write-Output "WARN $w"}
foreach($e in $errors){Write-Output "ERROR $e"}
Write-Output ("Documents={0} Links={1} Paths={2} Commands={3} LegacySources={4} Errors={5}" -f $counts.documents,$counts.links,$counts.paths,$counts.commands,$counts.sources,$errors.Count)
if($errors.Count){exit 1}
Write-Output 'PASS documentation checks only; business tests/codegen/security not executed.'
