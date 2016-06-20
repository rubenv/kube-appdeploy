package appdeploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

var CleanTypes = []string{
	"deployment",
	"service",
	"secret",
}

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
	contextName string
	client      *unversioned.Client
}

var _ Target = &KubernetesTarget{}

func NewKubernetesTarget(contextName string) *KubernetesTarget {
	return &KubernetesTarget{}
}

func (t *KubernetesTarget) Prepare() error {
	po := clientcmd.NewDefaultPathOptions()

	c, err := po.GetStartingConfig()
	if err != nil {
		return err
	}

	context, ok := c.Contexts[t.contextName]
	if !ok {
		names := make([]string, 0)
		for name, _ := range c.Contexts {
			names = append(names, name)
		}

		return fmt.Errorf("Unknown context: %s, should be one of: %s", t.contextName, strings.Join(names, ", "))
	}

	authinfo, ok := c.AuthInfos[context.AuthInfo]
	if !ok {
		return fmt.Errorf("Badly configured context, unknown auth: %s", context.AuthInfo)
	}

	cluster, ok := c.Clusters[context.Cluster]
	if !ok {
		return fmt.Errorf("Badly configured context, unknown cluster: %s", context.Cluster)
	}

	config := &restclient.Config{
		Host: cluster.Server,
		TLSClientConfig: restclient.TLSClientConfig{
			CAFile:   cluster.CertificateAuthority,
			CertFile: authinfo.ClientCertificate,
			KeyFile:  authinfo.ClientKey,
		},
	}

	client, err := unversioned.New(config)
	if err != nil {
		return err
	}

	t.client = client
	return nil
}

func (t *KubernetesTarget) Apply(m Manifest, data []byte) error {
	panic("not implemented")
}

func (t *KubernetesTarget) Cleanup(items []Manifest) error {
	panic("not implemented")
}
