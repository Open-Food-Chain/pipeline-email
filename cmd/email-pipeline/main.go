package main

import (
	"github.com/The-New-Fork/email-pipeline/pkg/pipeline"
	"github.com/unchainio/pkg/iferr"
	"github.com/unchainio/pkg/xconfig"
	"github.com/unchainio/pkg/xlogger"
)

func main() {
	// load config
	cfg := new(pipeline.Config)
	info := new(xconfig.Info)

	errs := xconfig.Load(
		cfg,
		xconfig.FromPathFlag("cfg", ""),
		xconfig.FromEnv(),
		xconfig.GetInfo(info),
	)

	// load logger
	log, err := xlogger.New(cfg.Logger)
	iferr.Exit(err)

	iferrlog, err := xlogger.New(&xlogger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		CallerDepth: 4,
	})
	iferr.Exit(err)

	iferr.Default, err = iferr.New(iferr.WithLogger(iferrlog))
	iferr.Exit(err)

	log.Printf("Attempted to load configs from %+v", info.Paths)
	iferr.Warn(errs)

	p := pipeline.New(cfg, log)

	err = p.Start()
	iferr.Exit(err)

	for <- p.StopChannel {

	}

}
