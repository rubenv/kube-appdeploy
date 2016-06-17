package appdeploy

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

type ManifestSource interface {
	Names() ([]string, error)
	Get(name string) (io.ReadCloser, error)
}

type FolderSource struct {
	Path string
}

func NewFolderSource(path string) *FolderSource {
	return &FolderSource{
		Path: path,
	}
}

func (s *FolderSource) Names() ([]string, error) {
	files, err := ioutil.ReadDir(s.Path)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			names = append(names, file.Name())
		}
	}

	return names, nil
}

func (s *FolderSource) Get(name string) (io.ReadCloser, error) {
	path := path.Join(s.Path, name)
	return os.Open(path)
}
