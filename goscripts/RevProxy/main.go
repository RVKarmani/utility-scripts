package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path"

	utils "goscripts.utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fatih/color"
)

func main() {
	presentWorkingDir, err := os.Getwd()
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	// Assets section
	fileName := utils.GetSsmPluginName()
	GoOsArch := utils.GetGoosArchString()
	// Download assets one level up
	fileDirectory := path.Join(path.Dir(presentWorkingDir), "assets", GoOsArch)

	err = utils.DownloadAssetsFromGitlab(fileName, fileDirectory, GoOsArch)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	environment, err := utils.AskEnvironment()
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	environmentInput, err := utils.GenerateAwsConfigFromEnvInput(*environment)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	localPort := flag.String("lp", "443", "Local Port to bind")
	remotePort := flag.String("rp", "443", "Remote Port to bind")
	flag.Parse()

	color.Yellow("Parameters\nInstance Id: %v\nLocal Port: %v\nRemote Port: %v", environmentInput.InstanceId, *localPort, *remotePort)

	sessionInput := &ssm.StartSessionInput{
		Target:       aws.String(environmentInput.InstanceId),
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters: map[string][]string{
			"portNumber":      {*remotePort},
			"localPortNumber": {*localPort},
		},
	}

	session, err := utils.StartSession(context.TODO(), environmentInput.Cfg, sessionInput)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	sessionJson, err := json.Marshal(session)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	paramsJson, err := json.Marshal(sessionInput)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	pluginPath := path.Join(fileDirectory, fileName)
	if err := utils.RunExternalCommand(pluginPath, string(sessionJson),
		environmentInput.Cfg.Region, "StartSession",
		string(paramsJson)); err != nil {
		color.Red("[err] %v", err.Error())
	}

	// delete session
	if err := utils.TerminateSession(context.TODO(), environmentInput.Cfg, &ssm.TerminateSessionInput{
		SessionId: session.SessionId,
	}); err != nil {
		color.Red("[err] %v", err.Error())
	}
}
