package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type OpCode int

var OPERATIONS = map[string]OpCode{
	"MoveByExt":  1,
	"MoveByName": 2,
}

const (
	CODEMOVEBYEXT  = 1
	CODEMOVEBYNAME = 2
)

func main() {
	operationName := flag.String("op", "", "Operation name")
	directoryName := flag.String("dir", "", "Directory name where the files of the given extensions will be stored.")
	typeExt := flag.String("ext", "", "Extension of the files that will be move to the given directory.")
	fileName := flag.String("name", "", "File name or pattern of the files that will be move to the given directory.")
	flag.Parse()

	var opcode OpCode = OPERATIONS[*operationName]
	switch opcode {
	case CODEMOVEBYEXT:
		moveByExtension(typeExt, directoryName)
	case CODEMOVEBYNAME:
		moveByName(fileName, directoryName)
	default:
		fmt.Println("Invalid operation.")
		displayValidOptions()
	}
}

func getAndDisplayCurrentWorkingDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting the current working directory:", err)
	}
	fmt.Println("The current working directory is:", cwd)
	return cwd
}

func moveByExtension(typeExt *string, directoryName *string) {
	if *directoryName == "" || *typeExt == "" {
		fmt.Println("Extension or directory name not provided")
		os.Exit(1)
	}

	cwd := getAndDisplayCurrentWorkingDirectory()
	fmt.Printf("Creating /%s and searching for files with %s extension in %s directory\n", *directoryName, *typeExt, cwd)

	chFiles := make(chan string)
	go getFilesNamesWithPattern(cwd, *typeExt, isFileAndHasExtension, chFiles)
	createDirectoryAndMoveFiles(cwd, typeExt, directoryName, chFiles)
}

func createDirectoryAndMoveFiles(cwd string, pattern *string, directoryName *string, chFiles chan string) {
	newDirectoryPath := filepath.Join(cwd, *directoryName)
	err := os.MkdirAll(newDirectoryPath, os.ModePerm)
	if err != nil {
		fmt.Println("Error trying to create the directory:", err)
	}
	done := make(chan bool)
	msg := fmt.Sprintf("\nAll files that fulfill the pattern %s in %s were moved.\n", *pattern, cwd)
	go showProgress(msg, done)
	for file := range chFiles {
		oldPath := filepath.Join(cwd, file)
		newPath := filepath.Join(newDirectoryPath, file)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			fmt.Println("Error moving", oldPath, "to new path", newPath)
		}
	}
	done <- true
}

func moveByName(fileName *string, directoryName *string) {
	fmt.Println("dir name:", *directoryName, "file name:", *fileName)
	if *directoryName == "" || *fileName == "" {
		fmt.Println("File name or directory name not provided")
		os.Exit(1)
	}
	cwd := getAndDisplayCurrentWorkingDirectory()
	fmt.Printf("Creating /%s and searching for files with %s name or similar in %s directory\n", *directoryName, *fileName, cwd)
	chFiles := make(chan string)
	go getFilesNamesWithPattern(cwd, *fileName, isFileAndFileNameContains, chFiles)
	createDirectoryAndMoveFiles(cwd, fileName, directoryName, chFiles)
}

func getFilesNamesWithPattern(cwd, pattern string, patternFn func(fs.DirEntry, string) bool, chFiles chan string) {
	defer close(chFiles)
	files, err := os.ReadDir(cwd)
	if err != nil {
		fmt.Println("Error trying to list the content of", cwd, ":", err)
		os.Exit(1)
	}
	for _, file := range files {
		if patternFn(file, pattern) {
			chFiles <- file.Name()
		}
	}
}

func isFileAndFileNameContains(file fs.DirEntry, pattern string) bool {
	return !file.IsDir() && strings.Contains(file.Name(), pattern)
}

func isFileAndHasExtension(file fs.DirEntry, extension string) bool {
	return !file.IsDir() && strings.HasSuffix(file.Name(), extension)
}

func showProgress(msgSuccess string, done chan bool) {
	for {
		select {
		case <-done:
			fmt.Print(msgSuccess)
			return
		default:
			fmt.Print(".")
			time.Sleep(100 * time.Millisecond)
		}
	}
}
