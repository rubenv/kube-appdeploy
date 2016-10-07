package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"k8s.io/client-go/1.4/rest"
	"k8s.io/client-go/1.4/tools/clientcmd"

	"github.com/rubenv/kube-appdeploy"
)

func main() {
	err := do()
	if err != nil {
		log.Fatal(err)
	}
}

func do() error {
	var target appdeploy.Target
	src := appdeploy.NewFolderSource(os.Args[1])

	/*
		folder := "/Users/ruben/Desktop/out"

		if folder == "" {
			log.Fatal("No output folder specified")
		}

		target = appdeploy.NewFolderTarget(folder)
	*/

	contextName := "vagrant-single"

	// Prepare Kubernetes client
	po := clientcmd.NewDefaultPathOptions()

	c, err := po.GetStartingConfig()
	if err != nil {
		return err
	}

	context, ok := c.Contexts[contextName]
	if !ok {
		names := make([]string, 0)
		for name, _ := range c.Contexts {
			names = append(names, name)
		}

		return fmt.Errorf("Unknown context: %s, should be one of: %s", contextName, strings.Join(names, ", "))
	}

	authinfo, ok := c.AuthInfos[context.AuthInfo]
	if !ok {
		return fmt.Errorf("Badly configured context, unknown auth: %s", context.AuthInfo)
	}

	cluster, ok := c.Clusters[context.Cluster]
	if !ok {
		return fmt.Errorf("Badly configured context, unknown cluster: %s", context.Cluster)
	}

	config := &rest.Config{
		Host:        cluster.Server,
		BearerToken: authinfo.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:   cluster.CertificateAuthority,
			CertFile: authinfo.ClientCertificate,
			KeyFile:  authinfo.ClientKey,
		},
	}

	target = appdeploy.NewKubernetesTarget(config)

	return appdeploy.Process(src, target)
}
