package internal

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	APIKey string   `json:"apiKey"`
	Ops    []OpJSON `json:"ops"`
}

func LoadConfig(filename string) (c Config, err error) {
	if filename == "" {
		filename = "notionutil.config.yml"
	}

	f, err := os.Open(filename)
	if err != nil {
		return c, err
	}

	defer func() {
		_ = f.Close()
	}()
	b, err := io.ReadAll(f)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(b, &c)
	return
}
