package main

import (
	"embed"
	"fmt"
	"github.com/jessevdk/go-flags"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed jar/*.jar
var jarDir embed.FS

type Option struct {
	JvmOpt string `short:"j" long:"jvmOpts" description:"JVM options for this app"`
}

func main() {
	var opt Option
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Printf("failed to parse options: %s\n", err)
	}
	jvmOpt := opt.JvmOpt
	fmt.Println(jvmOpt)
	files, err := jarDir.ReadDir("jar")
	if err != nil {
		fmt.Printf("failed to read embedded dir 'jar': %s\n", err)
	}
	var jarName string
	for i := range files {
		entry := files[i]
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".jar") {
			continue
		}
		jarName = name
		break
	}
	jarFile, err := jarDir.Open("jar/" + jarName)
	if err != nil {
		fmt.Printf("failed to open embedded file 'jar/%s': %s\n", jarName, err)
	}
	tmpJar, err := os.CreateTemp("./", "*.jar")
	tmpJarName := tmpJar.Name()
	defer func(name string) {
		jarFile.Close()
		tmpJar.Close()
		err := os.Remove(name)
		if err != nil {
			fmt.Printf("failed to remove temp file %s: %s\n", name, err)
		}
	}(tmpJarName)
	_, err = io.Copy(tmpJar, jarFile)
	if err != nil {
		fmt.Printf("failed to copy embedded file 'jar/%s': %s to ./temp.jar\n", jarName, err)
	}
	fmt.Println("jar file name: " + jarName)
	absJarPath, err := filepath.Abs(tmpJarName)
	jvmOpt = strings.TrimSpace(jvmOpt)
	var command *exec.Cmd
	if len(jvmOpt) > 0 {
		command = exec.Command("java", "-jar", jvmOpt, absJarPath)
	} else {
		command = exec.Command("java", "-jar", absJarPath)
	}
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		fmt.Printf("failed to run command --- %s\n", err)
	}
}
