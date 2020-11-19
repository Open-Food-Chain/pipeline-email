package pipeline

import "github.com/unchainio/interfaces/logger"

type Pipeline struct {
	cfg         *Config
	log         logger.Logger
	stopChannel chan bool
}

func New(cfg *Config, log logger.Logger) *Pipeline {
	return &Pipeline{
		cfg:         cfg,
		log:         log,
		stopChannel: make(chan bool),
	}
}
