package appdeploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"

	"github.com/rubenv/kube-appdeploy/kubectl"
)

var CleanTypes = []string{
	"deployment",
	"service",
}

type Target interface {
	Prepare(vars *ProcessVariables) error
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

func (t *FolderTarget) Prepare(vars *ProcessVariables) error {
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
	config    *rest.Config
	client    *kubernetes.Clientset
	kubectl   *kubectl.KubeCtl
	namespace string
}

var _ Target = &KubernetesTarget{}

func NewKubernetesTarget(config *rest.Config) *KubernetesTarget {
	return &KubernetesTarget{
		config: config,
	}
}

func (t *KubernetesTarget) Prepare(vars *ProcessVariables) error {
	client, err := kubernetes.NewForConfig(t.config)
	if err != nil {
		return err
	}

	t.client = client

	// Copy some vars
	t.namespace = vars.Namespace
	t.kubectl = kubectl.NewKubeCtl(t.config, t.namespace)

	// Ensure we have the needed namespace
	nsClient := t.client.Core().Namespaces()

	create := false
	_, err = nsClient.Get(t.namespace)
	if err != nil {
		ignore := false
		if e, ok := err.(*errors.StatusError); ok {
			if e.ErrStatus.Reason == "NotFound" {
				ignore = true
				create = true
			}
		}
		if !ignore {
			return err
		}
	}
	if create {
		_, err = nsClient.Create(&v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: t.namespace,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *KubernetesTarget) Apply(m Manifest, data []byte) error {
	_, err := t.runKubeCtl(data, "apply", "-f", "-")
	return err
}

func (t *KubernetesTarget) Cleanup(items []Manifest) error {
	for _, ct := range CleanTypes {
		err := t.cleanType(items, ct)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *KubernetesTarget) cleanType(items []Manifest, ct string) error {
	out, err := t.runKubeCtl(nil, "get", ct, "-o", "name")
	if err != nil {
		return err
	}

	known := []string{}
	for _, m := range items {
		if strings.ToLower(m.Kind) == ct {
			known = append(known, fmt.Sprintf("%s/%s", ct, m.Metadata.Name))
		}
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		found := false
		for _, k := range known {
			if line == k {
				found = true
			}
		}

		if !found {
			_, err := t.runKubeCtl(nil, "delete", line)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *KubernetesTarget) runKubeCtl(stdin []byte, args ...string) (string, error) {
	return t.kubectl.Run(stdin, args...)
}
