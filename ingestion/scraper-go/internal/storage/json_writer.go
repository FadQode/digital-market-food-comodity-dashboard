package storage

import (
	"os"
)

func SaveJSON(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
