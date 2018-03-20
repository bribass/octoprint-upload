package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jehiah/go-strftime"
	"gopkg.in/gcfg.v1"
)

const ConfigFileName = "opu.conf"
const ConfigEnvVar = "OPU_CONFIG"

func main() {
	// Configuration
	config := struct {
		Config struct {
			Url            string
			ApiKey         string
			UploadDir      string
			UploadToSdcard bool
		}
	}{}
	configFile, envPresent := os.LookupEnv(ConfigEnvVar)
	if !envPresent {
		configFile = ConfigFileName
	}
	err := gcfg.ReadFileInto(&config, configFile)
	if err != nil {
		log.Fatalf("Failed to parse config file '%s': %s", configFile, err)
	}
	// Sanity check config
	_, err = url.ParseRequestURI(config.Config.Url)
	if err != nil {
		log.Fatalf("Octoprint URL in config file is invalid: %s", err)
	}
	if len(config.Config.ApiKey) == 0 {
		log.Fatal("API Key in config file is missing")
	}
	fmt.Println("url =", config.Config.Url)
	fmt.Println("apikey =", config.Config.ApiKey)
	fmt.Println("upload dir =", config.Config.UploadDir)
	fmt.Println("sdcard =", config.Config.UploadToSdcard)

	// Compose the full URL of the upload
	uploadLocation := "local"
	if config.Config.UploadToSdcard {
		uploadLocation = "sdcard"
	}
	currentTime := time.Now()
	uploadDir := strftime.Format(config.Config.UploadDir, currentTime)
	uploadUrl := fmt.Sprintf("%s/api/files/%s", config.Config.Url, uploadLocation)
	fmt.Println("full upload =", uploadUrl)

	// Create the HTTP client
	client := http.Client{}

	// Create the directory
	var bodyBuf bytes.Buffer
	body := multipart.NewWriter(&bodyBuf)
	SetFormField(body, "foldername", uploadDir)
	body.Close()
	req, err := http.NewRequest("POST", uploadUrl, &bodyBuf)
	req.Header.Set("X-Api-Key", config.Config.ApiKey)
	req.Header.Set("Content-Type", body.FormDataContentType())
	resp, err := client.Do(req)
	fmt.Println("mkdir response = ")
	resp.Write(os.Stdout)

	// Upload the files
	for _, file := range os.Args[1:] {
		bodyBuf.Reset()
		body = multipart.NewWriter(&bodyBuf)
		SetFormField(body, "select", "false")
		SetFormField(body, "print", "false")
		SetFormField(body, "path", uploadDir)
		SetFormFile(body, "file", file)
		req, err = http.NewRequest("POST", uploadUrl, &bodyBuf)
		req.Header.Set("X-Api-Key", config.Config.ApiKey)
		req.Header.Set("Content-Type", body.FormDataContentType())
		resp, err = client.Do(req)
		fmt.Println("upload response = ")
		resp.Write(os.Stdout)
	}
}

func SetFormField(body *multipart.Writer, field, value string) {
	formField, _ := body.CreateFormField(field)
	_, _ = formField.Write([]byte(value))
}

func SetFormFile(body *multipart.Writer, field, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	fw, err := body.CreateFormFile(field, filename)
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}
}

