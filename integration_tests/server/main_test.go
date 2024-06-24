package server

import (
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {

	code := m.Run()
	os.Exit(code)
}

func TestExample(t *testing.T) {
	t.Fatal("test error")
}
