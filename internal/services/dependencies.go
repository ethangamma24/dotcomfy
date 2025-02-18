package services

import (
	"fmt"
	"os"
	"path/filepath"

	Config "dotcomfy/internal/config"
)

func InstallDependenciesLinux(config Config.Config) error {
	cfg_dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dotcomfy_dir := filepath.Join(cfg_dir, "dotcomfy")
	dependencies := config.Dependencies
	fmt.Fprintf(os.Stderr, "DEBUGPRINT: dependencies.go:17: dependencies=%+v\n", dependencies)
	package_manager, err := checkPackageManager()
	fmt.Println(dotcomfy_dir)

	for dependency := range dependencies {
		dependency_map, err := Config.GetDependency(dependency)

		if err != nil {
			return err
		}

		fmt.Println(dependency)
		fmt.Println(dependency_map)
		fmt.Println(len(dependency_map))

		// TODO: handle all fields being empty means it should just be installed at latest version from package manager
		if len(dependency_map) == 0 {
			fmt.Printf("Installing %s from package manager %s...\n", dependency, package_manager)
			err = installPackage(package_manager, dependency, "")
			if err != nil {
				fmt.Println("Error installing package:", err)
			}
			continue
		}
		for k, v := range dependency_map {
			switch k {
			case "version":
				fmt.Printf("Installing %s at version %s from package manager %s...\n", dependency, v.(string), package_manager)
				err = installPackage(package_manager, dependency, v.(string))
				if err != nil {
					fmt.Println("Error installing package:", err)
				}
			/*
				case "steps":
					for i, step := range v.([]interface{}) {
						fmt.Printf("Step %d: %s\n", i, step)
					}
				case "post_install_steps":
					for i, step := range v.([]interface{}) {
						fmt.Printf("Post install Step %d: %s\n", i, step)
					}
				case "script":
					fmt.Printf("Script location: %s\n", v)
					file_path := filepath.Join(dotcomfy_dir, v.(string))
					_, err := os.Stat(file_path)
					if os.IsNotExist(err) {
						fmt.Printf("File %s does not exist. Ensure it's in the directory %s\n", v, dotcomfy_dir)
						continue
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "DEBUGPRINT: dependencies.go:41: err=%+v\n", err)
						continue
					}
					content, err := os.ReadFile(file_path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "DEBUGPRINT: dependencies.go:46: err=%+v\n", err)
						continue
					}
					fmt.Printf("Script content: %s\n", string(content))
				case "post_install_script":
					fmt.Printf("Post install Script location: %s\n", v)
					file_path := filepath.Join(dotcomfy_dir, v.(string))
					_, err := os.Stat(file_path)
					if os.IsNotExist(err) {
						fmt.Printf("File %s does not exist. Ensure it's in the directory %s\n", v, dotcomfy_dir)
						continue
					}
					if err != nil {
						fmt.Fprintf(os.Stderr, "DEBUGPRINT: dependencies.go:59: err=%+v\n", err)
						continue
					}
					content, err := os.ReadFile(file_path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "DEBUGPRINT: dependencies.go:69: err=%+v\n", err)
						continue
					}
					fmt.Printf("Post Install Script content: %s\n", string(content))
			*/
			default:
				fmt.Printf("Unknown key: %s\n", k)
			}
		}

		if len(dependency_map) == 0 {
			fmt.Println("Installing package at latest version from package manager")
		}
	}
	return nil
}
