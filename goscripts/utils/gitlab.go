package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// Template format - OsArch String, version, fileName
const PROJECT_ID = "[PROJECT_ID]"
const PACKAGE_REGISTRY_URL_TEMPLATE = "https://gitlab.com/api/v4/projects/%s/packages/generic/%s/%s/%s"
const ACCESS_TOKEN = "[ACCESS_TOKEN]"
const PACKAGE_VERSION = "[PACKAGE_VERSION]"

// Wrapper for gitlab specific stuff
func DownloadAssetsFromGitlab(fileName string, fileDirectory string, goOsArch string) error {
	PACKAGE_REGISTRY_URL := fmt.Sprintf(PACKAGE_REGISTRY_URL_TEMPLATE, PROJECT_ID, goOsArch, PACKAGE_VERSION, fileName)
	return AssetDownloader(fileName, fileDirectory, PACKAGE_REGISTRY_URL, ACCESS_TOKEN)
}

func AssetDownloader(fileName string, fileDirectory string, url string, token string) error {
	err := os.MkdirAll(fileDirectory, os.ModePerm)
	if err != nil {
		return err
	}

	filePath := path.Join(fileDirectory, fileName)

	// Check if downloaded before
	if fileCheck(filePath) {
		color.Yellow("File %s exists in path, no need to download", fileName)
		return nil
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	err = os.Chmod(filePath, 0777)
	if err != nil {
		return err
	}

	// Get the data
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("PRIVATE-TOKEN", token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Bad response code")
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		fmt.Sprintf("Downloading %s", fileName),
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	// _, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func fileCheck(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
