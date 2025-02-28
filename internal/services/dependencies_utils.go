package services

import (
	"errors"
	"fmt"
	"log"
	"os/exec"

	Config "dotcomfy/internal/config"
)

/*
 * This is super janky right now because there aren't any go libraries that
 * abstract away package management for every major package manager. I plan
 * to get this just working in the meantime, then go back and contribute to
 * @REF [syspkg](https://github.com/bluet/syspkg) to fill out the missing
 * package managers.
 */

func CheckPackageManager() (string, error) {
	exists := func(pm string) bool {
		_, err := exec.LookPath(pm)
		return err == nil
	}

	if exists("apt") {
		return "apt", nil
	} else if exists("dnf") {
		return "dnf", nil
	} else if exists("yum") {
		return "yum", nil
	} else if exists("yay") {
		return "yay", nil
	} else if exists("pacman") {
		return "pacman", nil
	} else if exists("zypper") {
		return "zypper", nil
	} else {
		return "", errors.New("Unknown package manager")
	}
}

func InstallDependency(d *Config.Dependency, pm string) []error {
	var needs []string
	var errs []error

	needs = d.Needs
	if needs != nil {
		for _, need := range needs {
			n, error := Config.GetDependency(need)
			if error != nil {
				fmt.Println(error)
				err := errors.New("Error getting dependency \"" + need + "\"...")
				fmt.Println(err)
				errs = append(errs, err)
				return errs
			}
			if n.FailedInstall {
				err := errors.New("Dependency \"" + need + "\" previously failed to install, skipping \"" + d.Name + "\"...")
				fmt.Println(err)
				errs = append(errs, err)
				return errs
			}
			err := InstallDependency(n, pm)
			if err != nil {
				errs = append(errs, err...)
			}
		}
	}

	if d.Installed {
		return errs
	} else if d.FailedInstall {
		err := errors.New("Dependency \"" + d.Name + "\" previously failed to install, skipping...")
		fmt.Println(err)
		errs = append(errs, err)
		return errs
	} else if d.Version != "" {
		err := InstallPackage(pm, d.Name, d.Version)
		if err != nil {
			d.FailedInstall = true
			fmt.Println("Dependency \"" + d.Name + "\" failed to install from package manager...")
			errs = append(errs, err)
		}
		if d.PostInstallSteps != nil {
			err := HandleSteps(d.PostInstallSteps)
			if err != nil {
				d.FailedInstall = true
				fmt.Println("Dependency \"" + d.Name + "\" failed during the post install steps...")
				errs = append(errs, err)
				return errs
			}
		} else if d.PostInstallScript != "" {
			// TODO: Handle post install script
		}
		d.Installed = true
	} else {
		fmt.Println("Installing dependency \"" + d.Name + "\"...")
		if d.Steps != nil {
			err := HandleSteps(d.Steps)
			if err != nil {
				d.FailedInstall = true
				fmt.Println("Dependency \"" + d.Name + "\" failed during the install steps...")
				errs = append(errs, err)
				return errs
			}
		} else {
			// TODO: Handle script
		}
		d.Installed = true
	}
	return errs
}

func InstallPackage(pm string, pkg string, version string) error {
	switch pm {
	case "apt":
		if version != "" {
			pkg = pkg + "=" + version
		}
		err := exec.Command("sudo", "apt", "install", "-y", pkg).Run()
		return err
	case "dnf":
		if version != "" {
			pkg = pkg + "-" + version
		}
		cmd := fmt.Sprintf("sudo -S dnf install %s -y --skip-unavailable", pkg)
		command := exec.Command("/bin/sh", "-c", cmd)
		_, err := command.CombinedOutput()
		return err
	case "yum":
		if version != "" {
			pkg = pkg + "=" + version
		}
		err := exec.Command("sudo", "yum", "install", "-y", pkg).Run()
		return err
	case "pacman":
		if version != "" {
			pkg = pkg + "=" + version
		}
		cmd := fmt.Sprintf("sudo -S pacman -S %s --noconfirm", pkg)
		command := exec.Command("/bin/sh", "-c", cmd)
		_, err := command.CombinedOutput()
		return err
	case "yay":
		if version != "" {
			log.Output(1, "Version not supported for yay")
		}
		err := exec.Command("yay", "--noconfirm", pkg).Run()
		return err
	case "zypper":
		if version != "" {
			pkg = pkg + "=" + version
		}
		err := exec.Command("sudo", "zypper", "install", "-y", pkg).Run()
		return err
	default:
		return errors.New("Unknown package manager")
	}
}

func HandleSteps(steps []string) error {
	for _, step := range steps {
		cmd := exec.Command("/bin/sh", "-c", step)
		output, err := cmd.CombinedOutput()
		fmt.Println(string(output))
		if err != nil {
			return err
		}
	}
	return nil
}
