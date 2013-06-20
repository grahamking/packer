package vmware

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// A driver is able to talk to VMware, control virtual machines, etc.
type Driver interface {
	// CreateDisk creates a virtual disk with the given size.
	CreateDisk(string, string) error

	// Checks if the VMX file at the given path is running.
	IsRunning(string) (bool, error)

	// Start starts a VM specified by the path to the VMX given.
	Start(string) error

	// Stop stops a VM specified by the path to the VMX given.
	Stop(string) error

	// Verify checks to make sure that this driver should function
	// properly. This should check that all the files it will use
	// appear to exist and so on. If everything is okay, this doesn't
	// return an error. Otherwise, this returns an error.
	Verify() error
}

// Fusion5Driver is a driver that can run VMWare Fusion 5.
type Fusion5Driver struct {
	// This is the path to the "VMware Fusion.app"
	AppPath string
}

func (d *Fusion5Driver) CreateDisk(output string, size string) error {
	cmd := exec.Command(d.vdiskManagerPath(), "-c", "-s", size, "-a", "lsilogic", "-t", "1", output)
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) IsRunning(vmxPath string) (bool, error) {
	vmxPath, err := filepath.Abs(vmxPath)
	if err != nil {
		return false, err
	}

	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "list")
	stdout, _, err := d.runAndLog(cmd)
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(stdout, "\n") {
		if line == vmxPath {
			return true, nil
		}
	}

	return false, nil
}

func (d *Fusion5Driver) Start(vmxPath string) error {
	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "start", vmxPath, "gui")
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) Stop(vmxPath string) error {
	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "stop", vmxPath, "hard")
	if _, _, err := d.runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) Verify() error {
	if _, err := os.Stat(d.AppPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Fusion application not found at path: %s", d.AppPath)
		}

		return err
	}

	if _, err := os.Stat(d.vmrunPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Critical application 'vmrun' not found at path: %s", d.vmrunPath())
		}

		return err
	}

	if _, err := os.Stat(d.vdiskManagerPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Critical application vdisk manager not found at path: %s", d.vdiskManagerPath())
		}

		return err
	}

	return nil
}

func (d *Fusion5Driver) vdiskManagerPath() string {
	return filepath.Join(d.AppPath, "Contents", "Library", "vmware-vdiskmanager")
}

func (d *Fusion5Driver) vmrunPath() string {
	return filepath.Join(d.AppPath, "Contents", "Library", "vmrun")
}

func (d *Fusion5Driver) runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing: %s %v", cmd.Path, cmd.Args[1:])
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	log.Printf("stdout: %s", strings.TrimSpace(stdout.String()))
	log.Printf("stderr: %s", strings.TrimSpace(stderr.String()))

	return stdout.String(), stderr.String(), err
}
