package config_test

import (
	"os"
	"testing"

	"github.com/DIEALOG/DIEALOG-backend/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnvReadsEnvironmentVariables(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	t.Setenv("APP_ENV", "test")
	t.Setenv("SERVER_ADDRESS", ":9999")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("INTERNAL_CRYPTO_URL", "http://crypto.example.com:8080")

	config.NewEnv()

	assert.Equal(t, "test", config.E.AppEnv)
	assert.Equal(t, ":9999", config.E.ServerAddress)
	assert.Equal(t, "db.example.com", config.E.DBHost)
	assert.Equal(t, "http://crypto.example.com:8080", config.E.InternalCryptoUrl)
}

func TestNewEnvReadsDotEnvAndPrefersEnvironment(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	require.NoError(t, os.WriteFile(".env", []byte(
		"SERVER_ADDRESS=:7777\nDB_HOST=file-db\n",
	), 0o600))

	t.Setenv("SERVER_ADDRESS", "")
	t.Setenv("DB_HOST", "env-db")

	config.NewEnv()

	assert.Equal(t, ":7777", config.E.ServerAddress)
	assert.Equal(t, "env-db", config.E.DBHost)
}
