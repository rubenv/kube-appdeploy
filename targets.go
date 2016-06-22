package appdeploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
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
	config    *restclient.Config
	client    *unversioned.Client
	namespace string
}

var _ Target = &KubernetesTarget{}

func NewKubernetesTarget(config *restclient.Config) *KubernetesTarget {
	return &KubernetesTarget{
		config: config,
	}
}

func (t *KubernetesTarget) Prepare(vars *ProcessVariables) error {
	client, err := unversioned.New(t.config)
	if err != nil {
		return err
	}

	t.client = client

	// Copy some vars
	t.namespace = vars.Namespace

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
	args = append(t.configArgs(), args...)

	cmd := exec.Command("kubectl", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (t *KubernetesTarget) configArgs() []string {
	args := []string{
		"--namespace", t.namespace,
	}

	cfg := t.config
	if cfg.Host != "" {
		args = append(args, "--server", cfg.Host)
	}
	if cfg.CAFile != "" {
		args = append(args, "--certificate-authority", cfg.CAFile)
	}
	if cfg.CertFile != "" {
		args = append(args, "--client-certificate", cfg.CertFile)
	}
	if cfg.CertFile != "" {
		args = append(args, "--client-key", cfg.KeyFile)
	}
	if cfg.BearerToken != "" {
		args = append(args, "--token", cfg.BearerToken)
	}

	return args
}
