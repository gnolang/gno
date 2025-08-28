package main

import "go.uber.org/zap"

type cmdLogger struct {
	logger *zap.Logger
}

func newCommandLogger(logger *zap.Logger) *cmdLogger {
	return &cmdLogger{
		logger: logger,
	}
}

func (c *cmdLogger) Info(msg string, args ...interface{}) {
	if len(args) == 0 {
		c.logger.Info(msg)

		return
	}

	c.logger.Info(msg, zap.Any("args", args))
}

func (c *cmdLogger) Debug(msg string, args ...interface{}) {
	if len(args) == 0 {
		c.logger.Debug(msg)

		return
	}

	c.logger.Debug(msg, zap.Any("args", args))
}

func (c *cmdLogger) Error(msg string, args ...interface{}) {
	if len(args) == 0 {
		c.logger.Error(msg)

		return
	}

	c.logger.Error(msg, zap.Any("args", args))
}
