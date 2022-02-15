// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os"
	"testing"
)

type example struct {
	Name       string
	Env        []string
	OnlyLocked bool
	SkipRemove bool
}

func (et example) Check(t *testing.T) {
	env := BuildEnv(t)
	logger := Logger{}
	vendir := Vendir{t, env.BinaryPath, logger}

	logger.Section(et.Name, func() {
		err := et.check(t, vendir)
		if err != nil {
			t.Fatalf("[example: %s] %s", et.Name, err)
		}
	})
}

func (et example) check(t *testing.T, vendir Vendir) error {
	tmpDir, err := os.MkdirTemp("", "vendir-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	_, _, err = execGit([]string{"clone", ".", tmpDir}, "../..")
	if err != nil {
		t.Fatalf("failed to copy repo to temp dir: %v", err)
	}

	dir := "examples/" + et.Name
	path := tmpDir + "/" + dir

	vendorPath := path + "/vendor"

	vendorDir, err := os.Stat(vendorPath)
	if err != nil {
		return fmt.Errorf("Expected no err for stat: %v", err)
	}
	if !vendorDir.IsDir() {
		return fmt.Errorf("Expected to be dir")
	}

	// remove all vendored bits
	if !et.SkipRemove {
		err = os.RemoveAll(vendorPath)
		if err != nil {
			return fmt.Errorf("Expected no err for remove all")
		}
	}

	if !et.OnlyLocked {
		_, err = vendir.RunWithOpts([]string{"sync"}, RunOpts{Dir: path, Env: et.Env})
		if err != nil {
			return fmt.Errorf("Expected no err for sync")
		}

		// This assumes that example's vendor directory is committed to git
		gitOut := gitDiffExamplesDir(t, dir)
		if gitOut != "" {
			return fmt.Errorf("Expected no diff, but was: >>>%s<<<", gitOut)
		}
	}

	_, err = vendir.RunWithOpts([]string{"sync", "--locked"}, RunOpts{Dir: path, Env: et.Env})
	if err != nil {
		return fmt.Errorf("Expected no err for sync locked")
	}

	gitOut := gitDiffExamplesDir(t, dir)
	if gitOut != "" {
		return fmt.Errorf("Expected no diff, but was: >>>%s<<<", gitOut)
	}

	return nil
}
