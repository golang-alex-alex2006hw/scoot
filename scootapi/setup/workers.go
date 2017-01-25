package setup

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/scootdev/scoot/os/temp"
	"github.com/scootdev/scoot/snapshot"
	"github.com/scootdev/scoot/snapshot/bundlestore"
	"github.com/scootdev/scoot/snapshot/git/gitdb"
	"github.com/scootdev/scoot/snapshot/git/repo"
)

// WorkersStrategy is a strategy to start workers and returns the config to pass to a scheduler to talk to them
type WorkersStrategy interface {
	// StartupWorkers should startup necessary workers, and returns the config to pass to Scheduler, or an error
	StartupWorkers() (string, error)
}

// In addition to count, we'll optionally want repoDir and storeHandle to initialize workers' gitdb.
// Whatever is unset will be given a default value.
type WorkerConfig struct {
	Count       int
	RepoDir     string
	StoreHandle string
}

const DefaultWorkerCount int = 5

// InMemoryWorkersStrategy will use in-memory workers (to test the Scheduler logic)
type InMemoryWorkersStrategy struct {
	workersCfg *WorkerConfig
}

// NewInMemoryWorkers creates a new InMemoryWorkersStartup
func NewInMemoryWorkers(workersCfg *WorkerConfig) *InMemoryWorkersStrategy {
	return &InMemoryWorkersStrategy{workersCfg: workersCfg}
}

func (s *InMemoryWorkersStrategy) StartupWorkers() (string, error) {
	log.Println("Using in-memory workers")
	if s.workersCfg.Count == 0 {
		return "local.memory", nil
	}

	return fmt.Sprintf(`{"Cluster": {"Type": "memory", "Count": %d}, "Workers": {"Type": "local"}}`, s.workersCfg.Count), nil
}

// LocalWorkersStrategy will startup workers running locally
type LocalWorkersStrategy struct {
	workersCfg *WorkerConfig
	builder    Builder
	cmds       *Cmds
	nextPort   int
}

// Start workers at 10100 for clarity
const firstWorkerPort = 10100

// NewLocalWorkers creates a new LocalWorkersStartup
func NewLocalWorkers(workersCfg *WorkerConfig, builder Builder, cmds *Cmds) *LocalWorkersStrategy {
	return &LocalWorkersStrategy{
		workersCfg: workersCfg,
		builder:    builder,
		cmds:       cmds,
		nextPort:   firstWorkerPort,
	}
}

func (s *LocalWorkersStrategy) StartupWorkers() (string, error) {
	log.Printf("Using local workers")
	if s.workersCfg.Count < 0 {
		return "", fmt.Errorf("LocalWorkers must start with at least 1 worker (or zero for default #): %v", s.workersCfg.Count)
	} else if s.workersCfg.Count == 0 {
		s.workersCfg.Count = DefaultWorkerCount
	}

	log.Printf("Using %d local workers", s.workersCfg.Count)

	bin, err := s.builder.Worker()
	if err != nil {
		return "", err
	}

	for i := 0; i < s.workersCfg.Count; i++ {
		httpPort := strconv.Itoa(s.nextPort)
		s.nextPort++
		thriftPort := strconv.Itoa(s.nextPort)
		s.nextPort++
		if err := s.cmds.Start(bin, "-thrift_addr", "localhost:"+thriftPort, "-http_addr", "localhost:"+httpPort,
			"-repo", s.workersCfg.RepoDir, "-bundlestore", s.workersCfg.StoreHandle,
		); err != nil {
			return "", err
		}
		if err := WaitForPort(httpPort); err != nil {
			return "", err
		}
	}

	return "local.local", nil
}

// Constructs a gitdb backed by repo-dir and using the store (filepath or url) for upload/download.
// If repoDir is not specified, the gitdb is backed by a tmp dir.
// If storeHandle is not specified, the gitdb will upload/download to a tmp dir.
func NewGitDB(tmpDir *temp.TempDir, repoDir, storeHandle string) (snapshot.DB, error) {
	// Make the repo, initializing from scratch if a git dir isn't provided.
	var r *repo.Repository
	var err error
	if repoDir == "" {
		if repoTmp, err := tmpDir.TempDir("worker_repo"); err != nil {
			return nil, err
		} else if r, err = repo.InitRepo(repoTmp.Dir); err != nil {
			return nil, err
		}
	} else {
		r, err = repo.NewRepository(repoDir)
	}

	// Make the store, backed by tmp dir if store location isn't provided.
	var s bundlestore.Store
	if storeHandle == "" {
		if storeDir, err := tmpDir.TempDir("worker_store"); err != nil {
			return nil, err
		} else {
			storeHandle = storeDir.Dir
		}
	}
	if !strings.HasPrefix(storeHandle, "http://") {
		if s, err = bundlestore.MakeFileStoreInTemp(&temp.TempDir{Dir: storeHandle}); err != nil {
			return nil, err
		}
	} else {
		s = bundlestore.MakeHTTPStore(storeHandle)
	}
	if s, err = bundlestore.MakeCachingBrowseStore(s, tmpDir); err != nil {
		return nil, err
	}

	// Make the db and convert it into a filer.
	db := gitdb.MakeDBFromRepo(r, tmpDir, nil, nil, &gitdb.BundlestoreConfig{Store: s}, gitdb.AutoUploadBundlestore)
	return db, nil
}
