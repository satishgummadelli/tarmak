// Copyright Jetstack Ltd. See LICENSE for details.
package e2e_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/jetstack/tarmak/pkg/version"
)

var (
	lastVersion = ""
)

func ensureTarmakBins(t *testing.T) {
	if lastVersion == "" {
		var err error
		lastVersion, err = version.LastVersion()
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := EnsureTarmakDownload("../../..", lastVersion); err != nil {
		t.Fatal(err)
	}
}

func TestAWSSingleCluster(t *testing.T) {
	ensureTarmakBins(t)
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		ti.binPath = fmt.Sprintf("../../../_output/tarmak")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	if err := ti.sonobuoy.RunAndVerify(); err != nil {
		t.Error(err)
	}
}

func TestAWSMultiCluster(t *testing.T) {
	ensureTarmakBins(t)
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = false
	ti.singleZone = false

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		ti.binPath = fmt.Sprintf("../../../_output/tarmak")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()
	t.Log("run hub apply command")
	c := ti.Command("--current-cluster", fmt.Sprintf("%s-hub", ti.environmentName), "cluster", "apply")
	// write error out to my stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	if err := ti.sonobuoy.RunAndVerify(); err != nil {
		t.Error(err)
	}
}

func TestAWSUpgradeTarmak(t *testing.T) {
	ensureTarmakBins(t)
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	ti.binPath = fmt.Sprintf("../../../tarmak_%s_%s_%s", lastVersion, runtime.GOOS, runtime.GOARCH)

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		ti.binPath = fmt.Sprintf("../../../_output/tarmak")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	ti.binPath = fmt.Sprintf("../../../_output/tarmak")

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	if err := ti.sonobuoy.RunAndVerify(); err != nil {
		t.Error(err)
	}
}

func TestAWSUpgradeKubernetes(t *testing.T) {
	ensureTarmakBins(t)
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		ti.binPath = fmt.Sprintf("../../../_output/tarmak")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	ti.UpdateKubernetesVersion()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	if err := ti.sonobuoy.RunAndVerify(); err != nil {
		t.Error(err)
	}
}
