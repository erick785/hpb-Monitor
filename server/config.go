package server

import (
	"errors"
	"fmt"
)

var DefaultConfig = &Config{}

type Config struct {
	NodeEndpoint string
	ScanEndpoint string
	HttpURL      string
	HttpPort     string
	StartBlock   int64
	EndBlock     int64
}

func (c *Config) Valid() error {
	if c.NodeEndpoint == "" {
		return errors.New("Config node endpoint is nil")
	}

	if c.ScanEndpoint == "" {
		return errors.New("Config scan endpoint is nil")
	}

	if c.StartBlock > c.EndBlock {
		return fmt.Errorf("start monitor args error ,startBlock %d,endBlock %d ", c.StartBlock, c.EndBlock)
	}
	return nil
}
