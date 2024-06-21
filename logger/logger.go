// Copyright (c) - Proprietary and confidential. Belongs to:
// - SOARES Lucas (lucas.soares.npro@gmail.com)
// - LADEUILLE Guillaume (guillaume.ladeuille@hotmail.fr)
// - MAZEYRIE Laetitia (laetitia.mazeyrie@gmail.com)
// - BELLARDIE Nicolas (nicolas.bellardie@gmail.com)
// All Rights Reserved. Unauthorized copying of this file, via any medium is strictly prohibited
// Written by SOARES Lucas <lucas.soares.npro@gmail.com>, September 2020

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

func init() {
	// ensure logger is not empty if setup has not been called
	Logger, _ = zap.NewProduction()
}

func setup(logLevel zapcore.Level) error {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)
	config.Encoding = "console"
	config.DisableStacktrace = true
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := config.Build()
	Logger = logger
	return err
}

func Close() {
	_ = Logger.Sync()
}

func Setup() error {
	return setup(zapcore.DebugLevel)
}
