package run

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Exec executes a golang string
func Exec(code string) (*string, error) {
	dir, err := ioutil.TempDir("", "goexec")
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "main.go")
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	_, err = file.WriteString(code)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("go", "run", path)
	bout, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	str := string(bout)
	return &str, nil
}
