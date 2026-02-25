# filename: .devcontainer/scripts/generate-selfsigned.ps1
New-Item -ItemType Directory -Force -Path ".\.devcontainer\certs" | Out-Null

$cert = New-SelfSignedCertificate `
  -DnsName "chat.local" `
  -CertStoreLocation "Cert:\CurrentUser\My" `
  -KeyExportPolicy Exportable `
  -NotAfter (Get-Date).AddYears(1)

$pwd = ConvertTo-SecureString -String "changeit" -Force -AsPlainText
Export-PfxCertificate -Cert "Cert:\CurrentUser\My\$($cert.Thumbprint)" -FilePath ".\.devcontainer\certs\tls.pfx" -Password $pwd | Out-Null

Write-Host "Created .devcontainer\certs\tls.pfx (password: changeit). Convert to tls.crt/tls.key with openssl if needed."
