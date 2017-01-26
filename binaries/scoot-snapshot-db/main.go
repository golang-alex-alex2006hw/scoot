package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/scootdev/scoot/os/temp"
	"github.com/scootdev/scoot/snapshot"
	"github.com/scootdev/scoot/snapshot/bundlestore"
	"github.com/scootdev/scoot/snapshot/cli"
	"github.com/scootdev/scoot/snapshot/git/gitdb"
	"github.com/scootdev/scoot/snapshot/git/repo"
)

func main() {
	inj := &injector{}
	cmd := cli.MakeDBCLI(inj)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

type injector struct {
	// dir that holds our bundles
	storeDir string
}

func (i *injector) RegisterFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVar(&i.storeDir, "bundlestore_path", "", "path to where we store bundles")
}

// If 'storeDir' is nil don't use bundlestore backend, else use a file-backed bundlestore at that location.
func (i *injector) Inject() (snapshot.DB, error) {
	tempDir, err := temp.TempDirDefault()
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dataRepo, err := repo.NewRepository(wd)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot create a repo in wd %v; scoot-snapshot-db must be run in a git repo: %v", wd, err)
	}

	if i.storeDir == "" {
		storeTmp, err := tempDir.TempDir("bundlestore")
		if err != nil {
			return nil, err
		}
		i.storeDir = storeTmp.Dir
	}

	store, err := bundlestore.MakeFileStore(i.storeDir)
	if err != nil {
		return nil, err
	}

	return gitdb.MakeDBFromRepo(dataRepo, tempDir, nil, nil, &gitdb.BundlestoreConfig{Store: store},
		gitdb.AutoUploadBundlestore), nil
}