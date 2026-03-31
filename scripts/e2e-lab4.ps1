$ErrorActionPreference = "Stop"

$base = "http://localhost:8080"

function Show-Status([int]$code) {
  Write-Host ("Status: " + $code)
}

Write-Host "1) guest GET /api/brake-pad-wear (expect 401)"
try {
  Invoke-RestMethod -Method Get -Uri ($base + "/api/brake-pad-wear") | Out-Null
  Write-Host "UNEXPECTED: success"
} catch {
  Show-Status ([int]$_.Exception.Response.StatusCode.value__)
}

Write-Host "2) login as demo (expect token)"
$loginBody = @{ username = "demo"; password = "demo" } | ConvertTo-Json
$login = Invoke-RestMethod -Method Post -Uri ($base + "/api/auth/login") -ContentType "application/json" -Body $loginBody
$tokenUser = $login.token
if (-not $tokenUser) { throw "No token returned for demo" }
$hUser = @{ Authorization = ("Bearer " + $tokenUser) }
Write-Host "demo token OK"

Write-Host "3) login as moderator"
$loginModBody = @{ username = "moderator"; password = "demo" } | ConvertTo-Json
$loginMod = Invoke-RestMethod -Method Post -Uri ($base + "/api/auth/login") -ContentType "application/json" -Body $loginModBody
$tokenMod = $loginMod.token
if (-not $tokenMod) { throw "No token returned for moderator" }
$hMod = @{ Authorization = ("Bearer " + $tokenMod) }
Write-Host "moderator token OK"

Write-Host "4) create draft by adding service_id=1"
$addBody = @{ service_id = 1; quantity = 1 } | ConvertTo-Json
$draft = Invoke-RestMethod -Method Post -Uri ($base + "/api/brake-pad-wear/cart/items") -Headers $hUser -ContentType "application/json" -Body $addBody

$draftId = $null
if ($draft.PSObject.Properties.Name -contains "ID") { $draftId = $draft.ID }
if (-not $draftId -and ($draft.PSObject.Properties.Name -contains "id")) { $draftId = $draft.id }
if (-not $draftId) { throw "Cannot detect draft id from response" }
Write-Host ("draft id=" + $draftId)

Write-Host "5) form application (status=formed) as creator"
$updBody = @{ status = "formed"; driving_style = "Агрессивный"; mileage = 25000 } | ConvertTo-Json
$formed = Invoke-RestMethod -Method Put -Uri ($base + "/api/brake-pad-wear/" + $draftId) -Headers $hUser -ContentType "application/json" -Body $updBody
Write-Host ("formed status=" + $formed.status)

Write-Host "6) creator list applications (expect array, count>=1)"
$listUser = Invoke-RestMethod -Method Get -Uri ($base + "/api/brake-pad-wear?status=formed") -Headers $hUser
Write-Host ("creator got " + $listUser.Count + " items")

Write-Host "7) moderator list applications (expect array, includes creator app)"
$listMod = Invoke-RestMethod -Method Get -Uri ($base + "/api/brake-pad-wear?status=formed") -Headers $hMod
Write-Host ("moderator got " + $listMod.Count + " items")

Write-Host "8) creator tries finish (expect 403)"
try {
  Invoke-RestMethod -Method Put -Uri ($base + "/api/brake-pad-wear/" + $draftId + "/finish") -Headers $hUser | Out-Null
  Write-Host "UNEXPECTED: finish success"
} catch {
  Show-Status ([int]$_.Exception.Response.StatusCode.value__)
}

Write-Host "9) moderator finishes (expect 200)"
$fin = Invoke-RestMethod -Method Put -Uri ($base + "/api/brake-pad-wear/" + $draftId + "/finish") -Headers $hMod
Write-Host ("finish status=" + $fin.Status)

Write-Host "10) logout moderator (blacklist)"
$lo = Invoke-RestMethod -Method Post -Uri ($base + "/api/auth/logout") -Headers $hMod
Write-Host ($lo.message)

Write-Host "11) use blacklisted token (expect 401)"
try {
  Invoke-RestMethod -Method Get -Uri ($base + "/api/brake-pad-wear?status=formed") -Headers $hMod | Out-Null
  Write-Host "UNEXPECTED: blacklisted still works"
} catch {
  Show-Status ([int]$_.Exception.Response.StatusCode.value__)
}

