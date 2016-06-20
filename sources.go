package appdeploy

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v1"
)

type ManifestSource interface {
	Names() ([]string, error)
	Get(name string) (io.ReadCloser, error)
	Variables() (*ProcessVariables, error)
}

var _ ManifestSource = &FolderSource{}

type FolderSource struct {
	Path      string
	variables *ProcessVariables
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
		name := file.Name()
		if name == "variables.yaml" {
			vars := NewProcessVariables()

			data, err := ioutil.ReadFile(path.Join(s.Path, name))
			if err != nil {
				return nil, err
			}

			err = yaml.Unmarshal(data, vars.Variables)
			if err != nil {
				return nil, err
			}

			s.variables = vars
			continue
		}

		if !file.IsDir() {
			names = append(names, name)
		}
	}

	return names, nil
}

func (s *FolderSource) Get(name string) (io.ReadCloser, error) {
	path := path.Join(s.Path, name)
	return os.Open(path)
}

func (s *FolderSource) Variables() (*ProcessVariables, error) {
	return s.variables, nil
}
