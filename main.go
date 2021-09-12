package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	CWD, _      = os.Getwd()
	source      string
	destination string
)

func init() {
	flag.StringVar(&source, "source", "", "File to zip")
	flag.StringVar(&destination, "output", "", "Output Zip file name")
	flag.Parse()
}

func main() {

	if source != "" && destination != "" {
		if err := buildAndZip(source, destination); err != nil {
			log.Fatal(err)
		}
		log.Println("Produced package:", destination)
	} else {
		log.Fatal("Properly use: golambda --source [file_name] --destination [file_name.zip]")
	}
}

func buildAndZip(file, dest string) error {

	// create a temporary directory
	tempDir, err := ioutil.TempDir("", fileNameWithoutExtension(file))
	if err != nil {
		log.Println(err)
		return err
	}

	// copy the source code file
	if _, err := copy(file, filepath.Join(tempDir, file)); err != nil {
		cleanup(tempDir)
		log.Println("Failed to copy source code file")
		return err
	}

	f, err := os.Create(filepath.Join(tempDir, "main.go"))
	if err != nil {
		cleanup(tempDir)
		log.Println("Failed to create boilerplate file")
		return err
	}

	defer f.Close()

	f.WriteString(getBoilerplate())
	f.Sync()

	CLI, err := exec.LookPath("go")
	if err != nil {
		cleanup(tempDir)
		log.Println("Failed to get GOPATH")
		return err
	}

	// run go mod init
	execute := exec.Cmd{
		Path: CLI,
		Args: []string{CLI, "mod", "init", "github.com/nhost.io/" + fileNameWithoutExtension(file)},
		Dir:  tempDir,
	}

	if err := execute.Run(); err != nil {
		cleanup(tempDir)
		log.Println("Failed to run go mod init")
		return err
	}

	// run go mod tidy
	execute = exec.Cmd{
		Path: CLI,
		Args: []string{CLI, "mod", "tidy"},
		Dir:  tempDir,
	}

	if err := execute.Run(); err != nil {
		cleanup(tempDir)
		log.Println("Failed to run go mod tidy")
		return err
	}

	// build the binary
	execute = exec.Cmd{
		Env:  []string{"GOOS=linux", "GOARCH=amd64"},
		Path: CLI,
		Args: []string{CLI, "build", "-o", filepath.Join(CWD, "main")},
		Dir:  tempDir,
	}
	execute.Env = append(execute.Env, os.Environ()...)

	if err := execute.Run(); err != nil {
		cleanup(tempDir)
		log.Println("Failed to build binary")
		return err
	}

	// zip the file
	if err := zipIt("main", dest); err != nil {
		cleanup(tempDir)
		log.Println("Failed to zip it")
		return err
	}

	// remove the binary
	if err := os.Remove("main"); err != nil {
		cleanup(tempDir)
		log.Println("Failed to remove binary")
		return err
	}

	return nil
}

func zipIt(source, location string) error {

	fileToZip, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = source

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	newZipFile, err := os.Create(location)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func cleanup(path string) {
	if err := os.Remove(path); err != nil {
		log.Println("Failed to remove: ", path)
	}
}

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func getBoilerplate() string {
	return `
package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	server := route(Handler)
	lambda.Start(server)
}

// route wraps echo server into Lambda Handler
func route(handler func(http.ResponseWriter, *http.Request)) func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		buf, _ := ioutil.ReadAll(request.Body)
		body := bytes.NewBuffer(buf)
		req, _ := http.NewRequest(request.HTTPMethod, request.Path, body)
		for k, v := range request.Headers {
			req.Header.Add(k, v)
		}

		q := request.URL.Query()
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		handler(rec, req)

		res := rec.Result()
		responseHeaders := make(map[string]string)
		for key, value := range res.Header {
			responseHeaders[key] = ""
			if len(value) > 0 {
				responseHeaders[key] = value[0]
			}
		}

		responseBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       err.Error(),
				Headers:    responseHeaders,
				StatusCode: http.StatusInternalServerError,
			}, err
		}

		responseHeaders["Access-Control-Allow-Origin"] = "*"
		responseHeaders["Access-Control-Allow-Headers"] = "origin,Accept,Authorization,Content-Type"

		return events.APIGatewayProxyResponse{
			Body:       string(responseBody),
			Headers:    responseHeaders,
			StatusCode: res.StatusCode,
		}, nil
	}
}`
}
