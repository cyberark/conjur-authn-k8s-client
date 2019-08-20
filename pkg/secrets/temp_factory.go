package secrets

import (
	"io/ioutil"
	"os"
	"strings"
)

const DEVSHM = "/dev/shm"

type TempFactory struct {
	path  string
	files []string
}

// Create a new temporary file factory.
// defer Cleanup() if you want the files removed.
func NewTempFactory(path string) TempFactory {
	if path == "" {
		path = DefaultTempPath()
	}
	return TempFactory{path: path}
}

// Default temporary file path
// Returns /dev/shm if it is a directory, otherwise home dir of current user
// Else returns the system default
func DefaultTempPath() string {
	fi, err := os.Stat(DEVSHM)
	if err == nil && fi.Mode().IsDir() {
		return DEVSHM
	}
	home, err := os.UserHomeDir()
	if err == nil {
		dir, _ := ioutil.TempDir(home, ".tmp")
		return dir
	}
	return os.TempDir()
}

// Create a temp file with given value. Returns the path.
func (tf *TempFactory) Push(bytes []byte) string {
	f, _ := ioutil.TempFile(tf.path, ".summon")
	defer f.Close()

	f.Write(bytes)
	name := f.Name()
	tf.files = append(tf.files, name)
	return name
}

// Remove the temporary files created with this factory.
func (tf *TempFactory) Cleanup() {
	for _, file := range tf.files {
		os.Remove(file)
	}
	// Also remove the tempdir if it's not DEVSHM
	if !strings.Contains(tf.path, DEVSHM) {
		os.Remove(tf.path)
	}
	tf = nil
}
