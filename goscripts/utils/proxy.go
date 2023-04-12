package utils

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const UAT string = "uat"
const PROD string = "prod"

const AWS_REGION = "[AWS_REGION]"

var environmentChoices = []string{
	UAT,
	PROD,
}

type EnvironmentInput struct {
	Cfg        aws.Config
	InstanceId string
}

func AskEnvironment() (*string, error) {
	prompt := &survey.Select{
		Message: "Choose environment to connect to inorder to set credentials:",
		Options: environmentChoices,
	}

	var envChoice string

	if err := survey.AskOne(prompt, &envChoice, survey.WithIcons(func(icons *survey.IconSet) {
		icons.SelectFocus.Format = "green+hb"
	}), survey.WithPageSize(5)); err != nil {
		return nil, err
	}

	return &envChoice, nil
}

type EnvironmentProps struct {
	awsAccessKey       string
	awsSecretAccessKey string
	awsInstanceId      string
}

var environmentPropsMap = map[string]EnvironmentProps{
	UAT:  {awsAccessKey: "UAT_ACCESS_KEY", awsSecretAccessKey: "UAT_SECRET_ACCESS_KEY", awsInstanceId: "UAT_AWS_INSTANCE"},
	PROD: {awsAccessKey: "PROD_ACCESS_KEY", awsSecretAccessKey: "PROD_SECRET_ACCESS_KEY", awsInstanceId: "PROD_AWS_INSTANCE"},
}

func GenerateAwsConfigFromEnvInput(envInput string) (*EnvironmentInput, error) {
	var environmentPropsRetr EnvironmentProps = environmentPropsMap[envInput]

	awsAccessKey := environmentPropsRetr.awsAccessKey
	awsSecretAccessKey := environmentPropsRetr.awsSecretAccessKey
	awsRegion := AWS_REGION
	awsInstanceId := environmentPropsRetr.awsInstanceId

	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}
	awsConfig.Region = awsRegion
	return &EnvironmentInput{Cfg: awsConfig, InstanceId: awsInstanceId}, nil
}

func RunExternalCommandInBackground(process string, args ...string) error {
	call := exec.Command(process, args...)
	call.Stderr = nil
	call.Stdout = nil
	call.Stdin = nil

	// ignore signal(sigint)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-sigs:
			case <-done:
				break
			}
		}
	}()
	defer close(done)

	// run subprocess
	if err := call.Start(); err != nil {
		log.Fatal(err)
	}
	return nil
}
