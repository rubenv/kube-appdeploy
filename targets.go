package appdeploy

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Target interface {
	Prepare() error
	Apply(m Manifest, data []byte) error
	Cleanup(items []Manifest) error
}

// ---------- Folder ----------

type FolderTarget struct {
	Path string
}

var _ Target = &FolderTarget{}

func NewFolderTarget(path string) *FolderTarget {
	return &FolderTarget{
		Path: path,
	}
}

func (t *FolderTarget) Prepare() error {
	return os.MkdirAll(t.Path, 0755)
}

func (t *FolderTarget) Apply(m Manifest, data []byte) error {
	return ioutil.WriteFile(m.Filename(t.Path), data, 0644)
}

func (t *FolderTarget) Cleanup(items []Manifest) error {
	files, err := ioutil.ReadDir(t.Path)
	if err != nil {
		return err
	}

	filenames := make([]string, 0, len(items))
	for _, item := range items {
		filenames = append(filenames, item.Filename(""))
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		sep := strings.Index(name, "--")
		if sep < 0 {
			continue
		}

		prefix := name[0:sep]
		found := false
		for _, t := range CleanTypes {
			if t == prefix {
				found = true
			}
		}
		if !found {
			continue
		}

		known := false
		for _, f := range filenames {
			if f == name {
				known = true
			}
		}

		if !known {
			err = os.Remove(path.Join(t.Path, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ---------- Kubernetes ----------

type KubernetesTarget struct {
}

var _ Target = &KubernetesTarget{}

func NewKubernetesTarget() *KubernetesTarget {
	return &KubernetesTarget{}
}

func (t *KubernetesTarget) Prepare() error {
	panic("not implemented")
}

func (t *KubernetesTarget) Apply(m Manifest, data []byte) error {
	panic("not implemented")
}

func (t *KubernetesTarget) Cleanup(items []Manifest) error {
	panic("not implemented")
}
