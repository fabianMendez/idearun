package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/antchfx/xmlquery"
)

var (
	projectDir string
	configName string
	list       bool
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&projectDir, "project-dir", wd, "project directory")
	flag.StringVar(&configName, "config", "", "configuration name")
	flag.BoolVar(&list, "list", false, "list configurations")
	flag.Parse()
}

func main() {
	fmt.Println("Project dir:", projectDir)
	workspacePath := filepath.Join(projectDir, ".idea/workspace.xml")
	workspaceFile, err := os.Open(workspacePath)
	if err != nil {
		log.Fatal(err)
	}
	defer workspaceFile.Close()

	root, err := xmlquery.Parse(workspaceFile)
	if err != nil {
		log.Fatal(err)
	}

	runManager := xmlquery.FindOne(root, `//project//component[@name="RunManager"]`)

	if !list {
		if configName == "" {
			selected := runManager.SelectAttr("selected")
			configName = selected[strings.Index(selected, ".")+1:]
		}
		fmt.Println("Configuration:", configName)
	}

	configurations := xmlquery.Find(runManager, `//configuration`)
	var configuration *xmlquery.Node

	for _, c := range configurations {
		name := strings.TrimSpace(c.SelectAttr("name"))
		if list {
			fmt.Println(name)
		} else {
			if name == configName {
				configuration = c
				break
			}
		}
	}

	if list {
		return
	}

	if configuration == nil {
		log.Fatal("configuration not found")
	}

	configType := configuration.SelectAttr("type")
	if configType == "GradleRunConfiguration" {
		var tasks []string
		taskNames := xmlquery.Find(configuration, `//option[@name="taskNames"]//list//option`)
		for _, taskName := range taskNames {
			tasks = append(tasks, taskName.SelectAttr("value"))
		}

		var env []string
		envEntries := xmlquery.Find(configuration, `//option[@name="env"]//map//entry`)
		for _, entry := range envEntries {
			env = append(env, fmt.Sprintf("%s=%s", entry.SelectAttr("key"), entry.SelectAttr("value")))
		}

		gradlePath := filepath.Join(projectDir, "gradlew")
		cmd := exec.Command(gradlePath, tasks...)
		// cmd := exec.Command("env")
		cmd.Dir = projectDir
		cmd.Env = append(cmd.Env, env...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		log.Println(cmd.String())
		_ = cmd.Run()
	} else {
		log.Fatal("unsupported configuration type: " + configType)
	}
}
