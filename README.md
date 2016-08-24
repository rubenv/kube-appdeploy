# kube-appdeploy

Work-in-progress deployment tool for Kubernetes-hosted applications.

## Getting started

First build the command line client. Make sure kubectl is available on your path and a cluster is running.

    go build bin/kube-appdeploy/main.go

The examples folder contains .yaml files to deploy nginx on your cluster. To do just that, run:

    ./main example

The appdeploy script will pick up all .yaml files in the folder and process these. 
You can use [text/template](https://godoc.org/text/template) syntax to change the computed content of the deploy configuration.
This example deploys 2 nodes in production and 1 node in all other environments:

        apiVersion: extensions/v1beta1
	kind: Deployment
	metadata:
	  name: my-nginx
	spec:
	{{ if eq .Variables.env "production" }}
	  replicas: 2
	{{ else }}
	  replicas: 1
	{{ end }}
	  template:
	    metadata:
	      labels:
		run: my-nginx
	    spec:
	      containers:
	      - name: my-nginx
		image: nginx
		ports:
		- containerPort: 80

The "env" variable is set in "variables.yaml" in yaml syntax as shown here:

    env: "development"


