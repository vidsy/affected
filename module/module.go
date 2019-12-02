package module

import (
	"encoding/json"
	"os/exec"
)

// ModFile is the structure of a go.mod file
type ModFile struct {
	Module struct {
		Path string
	}
}

// Path calls go mod edit -json to retrieve the current module path
func Path() (string, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")

	b, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var f ModFile
	if err := json.Unmarshal(b, &f); err != nil {
		return "", err
	}

	return f.Module.Path, nil
}
