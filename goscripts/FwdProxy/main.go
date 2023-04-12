package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path"
	"runtime"
	"strings"

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

	color.Yellow("Assets downloaded to %v", fileDirectory)

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

	localPort := flag.String("lp", "3128", "Local Port to bind")
	remotePort := flag.String("rp", "3128", "Remote Port to bind")
	flag.Parse()

	color.Yellow("Parameters\nInstance Id: %v\nLocal Port: %v\nRemote Port: %v", environmentInput.InstanceId, *localPort, *remotePort)

	pluginPath := path.Join(fileDirectory, fileName)
	proxySessionInput := &ssm.StartSessionInput{
		Target:       aws.String(environmentInput.InstanceId),
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters: map[string][]string{
			"portNumber":      {"3128"},
			"localPortNumber": {"3128"},
		},
	}

	proxySession, err := utils.StartSession(context.TODO(), environmentInput.Cfg, proxySessionInput)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	proxySessionJson, err := json.Marshal(proxySession)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	proxySessionParamsJson, err := json.Marshal(proxySessionInput)
	if err != nil {
		color.Red("[err] %v", err.Error())
	}

	// Launch browser
	if strings.ToLower(runtime.GOOS) == "windows" {
		if err := utils.RunExternalCommandInBackground("cmd.exe", "/c", "start", "chrome", "--proxy-server=localhost:3128", "--proxy-bypass-list=*localhost", "--temp-profile"); err != nil {
			color.Red("[err] %v", err.Error())
		}
	} else {

		if err := utils.RunExternalCommandInBackground("google-chrome", "--proxy-server=localhost:3128", "--proxy-bypass-list=*localhost", "--temp-profile"); err != nil {
			color.Red("[err] %v", err.Error())
		}
	}

	if err := utils.RunExternalCommand(pluginPath, string(proxySessionJson),
		environmentInput.Cfg.Region, "StartSession",
		string(proxySessionParamsJson)); err != nil {
		color.Red("[err] %v", err.Error())
	}

	// delete session
	if err := utils.TerminateSession(context.TODO(), environmentInput.Cfg, &ssm.TerminateSessionInput{
		SessionId: proxySession.SessionId,
	}); err != nil {
		color.Red("[err] %v", err.Error())
	}
}
