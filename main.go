package main

import (
	_ "embed"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var (

	//
	//	Embed the file containing our custom proxy code.
	//
	//go:embed proxy.go.tmpl
	proxy string

	//	Get the current working directory
	CWD, _ = os.Getwd()

	//	Golang function file to be built
	source string

	//	Destination zip file to be deployed on Lambda
	destination string

	//	Execution specific temporary directory to conveniently build our function
	tempDir string
)

const (
	goMod = "go.mod"
	goSum = "go.sum"
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
		} else {
			log.Println("Produced package:", destination)
		}

		//	Delete the temporary directory
		cleanup(tempDir)

	} else {
		log.Fatal("Proper use: golambda --source [file_name.go] --destination [file_name.zip]")
	}
}

func buildAndZip(file, dest string) error {

	//
	//	Create a temporary directory to store our built plugin in
	//

	var err error
	tempDir, err = ioutil.TempDir("", fileNameWithoutExtension(file))
	if err != nil {
		return err
	}

	//
	//	Copy the source code file to our temporary directory
	//

	if _, err := copy(file, filepath.Join(tempDir, file)); err != nil {
		log.Println("Failed to copy source code file")
		return err
	}

	//
	//	Copy go.mod and go.sum if they exist
	//

	gomodFiles := []string{goMod, goSum}
	for _, filename := range gomodFiles {
		if fileExists(filename) {
			if _, err := copy(filename, filepath.Join(tempDir, filename)); err != nil {
				log.Printf("Failed to copy %s file\n", filename)
				return err
			}
		}
	}

	//
	//	Create the main file which will have our proxy code
	//

	f, err := os.Create(filepath.Join(tempDir, "main.go"))
	if err != nil {
		log.Println("Failed to create boilerplate file")
		return err
	}

	defer f.Close()

	f.WriteString(proxy)
	f.Sync()

	//
	//	Fetch the Golang utility installation path
	//

	CLI, err := exec.LookPath("go")
	if err != nil {
		log.Println("Failed to get GOPATH")
		return err
	}

	if !fileExists(filepath.Join(tempDir, goMod)) {
		//
		//	Initialize the function package
		//

		execute := exec.Cmd{
			Path:   CLI,
			Args:   []string{CLI, "mod", "init", "github.com/nhost.io/" + fileNameWithoutExtension(file)},
			Dir:    tempDir,
			Stderr: os.Stderr,
			Stdout: os.Stdout,
		}

		if err := execute.Run(); err != nil {
			log.Println("Failed to run go mod init")
			return err
		}
	}

	//
	//	Download all the dependencies of our function
	//

	execute := exec.Cmd{
		Path:   CLI,
		Args:   []string{CLI, "mod", "tidy"},
		Dir:    tempDir,
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}

	if err := execute.Run(); err != nil {
		log.Println("Failed to run go mod tidy")
		return err
	}

	//
	//	Build the function binary
	//

	execute = exec.Cmd{
		Env:    []string{"GOOS=linux", "GOARCH=amd64"},
		Path:   CLI,
		Args:   []string{CLI, "build", "-o", filepath.Join(CWD, "main")},
		Dir:    tempDir,
		Stderr: os.Stderr,
		Stdout: os.Stdout,
	}
	execute.Env = append(execute.Env, os.Environ()...)

	if err := execute.Run(); err != nil {
		log.Println("Failed to build binary")
		return err
	}

	//
	//	Compress/zip the output file
	//

	if err := zipIt("main", dest); err != nil {
		log.Println("Failed to zip it")
		return err
	}

	//
	//	Delete the binary once our job is complete
	//

	if err := os.Remove("main"); err != nil {
		log.Println("Failed to remove binary")
		return err
	}

	return nil
}
