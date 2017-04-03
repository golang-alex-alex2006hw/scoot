package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/scootdev/scoot/common/log/hooks"
	"github.com/scootdev/scoot/fs/minfuse"
)

func main() {
	log.AddHook(hooks.NewContextHook())
	minfuse.SetupLog()
	if opts, err := minfuse.InitFlags(); err != nil {
		log.Info(err)
		return
	} else {
		minfuse.Runfs(opts)
	}
}
