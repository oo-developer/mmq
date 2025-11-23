package main

import (
	"fmt"
	"os"
	"path"

	"github.com/oo-developer/mmq/cli/module"
	api "github.com/oo-developer/mmq/pkg"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		printHelp()
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("[ERROR] Loading config: %s\n", err)
	}
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("[ERROR] Creating client : %v \n", err)
		os.Exit(1)
	}
	err = client.Connect()
	if err != nil {
		fmt.Printf("[ERROR] Connecting client : %v \n", err)
		os.Exit(1)
	}

	moduleName := args[0]
	commandName := args[1]

	if mod, ok := module.Modules[moduleName]; ok {
		err := mod.Execute(client, commandName, args[2:]...)
		if err != nil {
			fmt.Printf("[ERROR] %v\n", err)
			os.Exit(1)
		}
	}
}

func loadConfig() (*api.Config, error) {
	homeDir, _ := os.UserHomeDir()
	configFilePath := path.Join(homeDir, ".mmq")
	configFileName := path.Join(configFilePath, "config.json")
	err := os.MkdirAll(configFilePath, 0750)
	if err != nil {
		fmt.Printf("[ERROR] Creating config directory: %v\n", err)
		os.Exit(1)
	}
	config, err := api.LoadConfig(configFileName)
	if err != nil {
		fmt.Printf("[ERROR] Loading config file '%s' : %v \n", configFilePath, err)
		os.Exit(1)
	}
	config.ClientPrivateKeyFile = path.Join(configFilePath, "private_key.pem")
	return config, nil
}

func printHelp() {
	fmt.Printf("Usage: %s <module> <command> [options]\n\n", os.Args[0])
	fmt.Println("The modules are:")
	fmt.Printf("  %s users help\n", os.Args[0])
	fmt.Printf("  %s connections help\n", os.Args[0])
	os.Exit(0)
}
