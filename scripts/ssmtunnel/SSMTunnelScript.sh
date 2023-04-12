#!/usr/bin/env bash

RED="\e[1;31m"
GREEN="\e[1;32m"
ORANGE="\e[1;33m"
BLUE="\e[1;34m"
MAGENTA="\e[1;35m"

AWS_CONFIG_DIR="./.aws-config"
AWS_ZIP_FILE="awscliv2.zip"
SESSION_MANAGER_FILE="session-manager-plugin.deb"
AWS_PROFILE="sso-profile"

AWS_SSO_START_URL="[AWS_SSO_START_URL]"
AWS_SSO_ACCOUNT_ID="[AWS_SSO_ACCOUNT_ID]"
AWS_SSO_ROLE_NAME="sso-ssm-user-ps"
AWS_REGION="[AWS_REGION]"
AWS_OUTPUT="json"

setup_aws_tool() {
    if [[ $retValue -eq 0 ]]
    then
        conditional_curl "./$AWS_ZIP_FILE" "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" "$AWS_ZIP_FILE"
        unzip awscliv2.zip
        # sudo ./aws/install
        sudo ./aws/install -i /usr/local/aws-cli -b /usr/local/bin
    else
        print_color_echo "$GREEN" "No further setup required for AWS cli"
        print_color_echo "$GREEN" "$(aws --version) found"
    fi
}

setup_session_manager_plugin() {
    if [[ $retValue -eq $FAILURE ]]
    then
        conditional_curl "./$SESSION_MANAGER_FILE" "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" "$SESSION_MANAGER_FILE"
        sudo dpkg -i session-manager-plugin.deb
    else
        print_color_echo "$GREEN" "No further setup required for Session Manager Plugin"
    fi
}

# Arg format - Color code text
print_color_echo() {
    echo -e "$1$2 \e[0m"
}

# Arg format - command - returns 1 if present, 0 if not
check_command() {
    if ! command -v "$1" &> /dev/null
    then
        print_color_echo "$RED" "$1 not found"
        print_color_echo "$GREEN" "Setting up $1 tool"
        retValue=0
    else
        print_color_echo "$GREEN" "$1 found, skipping setup"
        retValue=1
    fi
    return "$retValue"
}

# Arg format - FilePath - URL - Output File Name
conditional_curl() {
    if [ -e "$1" ]
    then
        print_color_echo "$GREEN" "$1 exists no need to redownload."
    else
        curl "$2" -o "$3"
    fi
}

print_color_echo "$MAGENTA" "SSM Tunneling Utility"

retValue=1

# AWS CLI SETUP
print_color_echo "$BLUE" "\n[Step 1] Setting up AWS CLI"
check_command "aws"
retValue=$?
setup_aws_tool "$retValue"

# SSM Session Plugin
print_color_echo "$BLUE" "\n[Step 3] Setting up AWS Session Manager Plugin"
check_command "session-manager-plugin"
retValue=$?
setup_session_manager_plugin "$retValue"

# AWS SSO LOGIN
print_color_echo "$BLUE" "\n[Step 2] Setting up AWS SSO"
export AWS_CONFIG_FILE=$AWS_CONFIG_DIR/config
export AWS_SHARED_CREDENTIALS_FILE=$AWS_CONFIG_DIR/credentials

# Setting various properties of SSO profile
aws configure set profile.$AWS_PROFILE.sso_start_url $AWS_SSO_START_URL
aws configure set profile.$AWS_PROFILE.sso_region $AWS_REGION
aws configure set profile.$AWS_PROFILE.sso_account_id $AWS_SSO_ACCOUNT_ID
aws configure set profile.$AWS_PROFILE.sso_role_name $AWS_SSO_ROLE_NAME
aws configure set profile.$AWS_PROFILE.region $AWS_REGION
aws configure set profile.$AWS_PROFILE.output $AWS_OUTPUT

aws sso login --profile $AWS_PROFILE

COLOR_PROMPT=$(print_color_echo "$ORANGE" "Enter Instance ID to tunnel to: ")
read -r -p "$COLOR_PROMPT" INSTANCE_ID

COLOR_PROMPT=$(print_color_echo "$ORANGE" "Enter remote port to bind to: ")
read -r -p "$COLOR_PROMPT" REMOTE_PORT

RANDOM_PORT=$(comm -23 <(seq 49152 65535 | sort) <(ss -Htan | awk '{print $4}' | cut -d':' -f2 | sort -u) | shuf | head -n 1)
xdg-open http://localhost:"$RANDOM_PORT"

aws ssm start-session --profile $AWS_PROFILE --target "$INSTANCE_ID" --document-name AWS-StartPortForwardingSession --parameters "localPortNumber=$RANDOM_PORT,portNumber=$REMOTE_PORT"