Write-Host "SSM Tunneling Utility" -ForegroundColor Magenta

$AWS_CLI_TOOL_BIN = "/i https://awscli.amazonaws.com/AWSCLIV2.msi"
$SESSION_MANAGER_EXE_FILE = "SessionManagerPluginSetup.exe"
$SESSION_TOOL_BIN = "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/windows/$SESSION_MANAGER_EXE_FILE"
$TOOL_INSTALLED = $false

# SSO Properties
$AWS_PROFILE = "sso-profile"

# INPUTS
$AWS_SSO_START_URL = "[START_URL]"
$AWS_SSO_ACCOUNT_ID = "[AWS_ACCOUNT_ID]"
$AWS_SSO_ROLE_NAME = "sso-ssm-user-ps"
$AWS_REGION = "[AWS_REGION]"
$AWS_OUTPUT = "json"

function CheckToolNeedsInstall($command_str) {
  try {
    # Not directly executed in order to keep return intact
    Invoke-Expression $command_str
  }
  catch {
    return $true
  }
  return $false
}

function Get-ActiveTcpPort {
  # Use a hash set to avoid duplicates
  $portList = New-Object -TypeName Collections.Generic.HashSet[uint16]

  $properties = [Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties()

  $listener = $properties.GetActiveTcpListeners()
  $active = $properties.GetActiveTcpConnections()

  foreach ($serverPort in $listener) {
    [void]$portList.Add($serverPort.Port)
  }
  foreach ($clientPort in $active) {
    [void]$portList.Add($clientPort.LocalEndPoint.Port)
  }

  return $portList
}

function Get-InactiveTcpPort {
  [CmdletBinding()]
  Param(
    [Parameter(Position = 0)]
    [uint16]$Start = 1024,

    [Parameter(Position = 1)]
    [uint16]$End = 5000
  )
  $attempts = 100
  $counter = 0

  $activePorts = Get-ActiveTcpPort

  while ($counter -lt $attempts) {
    $counter++
    $port = Get-Random -Minimum ($Start -as [int]) -Maximum ($End -as [int])

    if ($port -notin $activePorts) {
      return $port
    }
  }
  $emsg = [string]::Format(
    'Unable to find available TCP Port. Range: {0}, Attempts: [{1}]',
    "[$Start - $End]",
    $attempts
  )
  throw $emsg
}

Write-Host "[Step 1] - AWS CLI V2 Setup" -ForegroundColor Blue
$AWS_TOOL_AVAIL = CheckToolNeedsInstall("aws --version")

if ($AWS_TOOL_AVAIL) {
  Write-Host "AWS CLI is not installed, starting the installation process"
  Start-Process msiexec.exe -ArgumentList $AWS_CLI_TOOL_BIN -Wait
  $TOOL_INSTALLED = $true
}
else {
  Write-Host "AWS Cli already installed, skipping setup" -ForegroundColor Yellow
}

Write-Host "[Step 2] - Session Manager Setup" -ForegroundColor Blue
$SESSION_TOOL_AVAIL = CheckToolNeedsInstall("session-manager-plugin --version")

if ($SESSION_TOOL_AVAIL) {

  Write-Host "Checking if $SESSION_MANAGER_EXE_FILE exists" -ForegroundColor Yellow
  $SESSION_FILE_CHECK = Test-Path -Path .\$SESSION_MANAGER_EXE_FILE -PathType Leaf

  if ($SESSION_FILE_CHECK) {
    Write-Host "$SESSION_MANAGER_EXE_FILE already exists, no need to download" -ForegroundColor Yellow
  }
  else {
    Write-Host "Downloading Session Manager Plugin installer" -ForegroundColor Yellow
    Invoke-WebRequest $SESSION_TOOL_BIN -OutFile $SESSION_MANAGER_EXE_FILE
  }
  Start-Process -FilePath .\$SESSION_MANAGER_EXE_FILE -Wait
  $TOOL_INSTALLED = $true
}
else {
  Write-Host "Session Manager Plugin already installed, skipping setup" -ForegroundColor Yellow
}

if ($TOOL_INSTALLED) {
  Write-Host "Atleast one tool has been installed as part of this script's execution" -ForegroundColor Yellow
  Write-Host "Resetting env PATH so that the tools are available within this execution otherwise we would need to restart PowerShell" -ForegroundColor Yellow
  $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User") 
}

Write-Host "[Step 3] - Setting up AWS Credentials" -ForegroundColor Blue

$env:AWS_CONFIG_FILE = ".\.aws-config\config"
$env:AWS_SHARED_CREDENTIALS_FILE = ".\aws-config\credentials"

# Remove-Item .\aws-config\* -Recurse -Force -ErrorAction SilentlyContinue
# New-Item .\aws-config -ItemType Directory -ErrorAction SilentlyContinue
# Set-Content .\aws-config\config $AWS_PROFILE

# Setting various properties of SSO profile
aws configure set profile.$AWS_PROFILE.sso_start_url $AWS_SSO_START_URL
aws configure set profile.$AWS_PROFILE.sso_region $AWS_REGION
aws configure set profile.$AWS_PROFILE.sso_account_id $AWS_SSO_ACCOUNT_ID
aws configure set profile.$AWS_PROFILE.sso_role_name $AWS_SSO_ROLE_NAME
aws configure set profile.$AWS_PROFILE.region $AWS_REGION
aws configure set profile.$AWS_PROFILE.output $AWS_OUTPUT

aws sso login --profile $AWS_PROFILE

Write-Host "[Step 4] - Starting tunneling process" -ForegroundColor Blue

$RANDOM_PORT = Get-InactiveTcpPort
$INSTANCE_ID = Read-Host -Prompt "Enter Instance ID to tunnel to"
$REMOTE_PORT = Read-Host -Prompt "Enter remote port to bind to"

# if ([string]::IsNullOrWhiteSpace($REMOTE_PORT)) {
#     $REMOTE_PORT = "22"
#  }

Write-Host "Remote Port is: $REMOTE_PORT" -ForegroundColor Yellow
Write-Host "Starting Tunnel session with $INSTANCE_ID on local Port $RANDOM_PORT and remote port $REMOTE_PORT" -ForegroundColor Yellow

Start-Process http://localhost:$RANDOM_PORT

aws ssm start-session --profile $AWS_PROFILE --target $INSTANCE_ID --document-name AWS-StartPortForwardingSession --parameters "localPortNumber=$RANDOM_PORT,portNumber=$REMOTE_PORT"
