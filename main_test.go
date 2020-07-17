package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigGetEnv(t *testing.T) {
	cfg := Config{}
	assert.Equal(t, EnvLocal, cfg.getEnv())

	cfg = Config{Env: EnvStaging}
	assert.Equal(t, EnvStaging, cfg.getEnv())

	cfg = Config{Env: EnvProduction}
	assert.Equal(t, EnvProduction, cfg.getEnv())
}
