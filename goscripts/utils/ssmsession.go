package utils

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/color"
)

func StartSession(ctx context.Context, awsConfig aws.Config, ssmInput *ssm.StartSessionInput) (*ssm.StartSessionOutput, error) {
	client := ssm.NewFromConfig(awsConfig)
	return client.StartSession(ctx, ssmInput)
}

func TerminateSession(ctx context.Context, awsConfig aws.Config, input *ssm.TerminateSessionInput) error {
	client := ssm.NewFromConfig(awsConfig)
	fmt.Printf("%s %s \n", color.YellowString("Delete Session"),
		color.YellowString(aws.ToString(input.SessionId)))
	_, err := client.TerminateSession(ctx, input)
	return err
}

func GetSsmPluginName() string {
	if strings.ToLower(runtime.GOOS) == "windows" {
		return "session-manager-plugin.exe"
	} else {
		return "session-manager-plugin"
	}
}
