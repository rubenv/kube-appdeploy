package main

import (
	"log"
	"os"

	"github.com/rubenv/kube-appdeploy"
)

func main() {
	var target appdeploy.Target
	src := appdeploy.NewFolderSource(os.Args[1])

	/*
		folder := "/Users/ruben/Desktop/out"

		if folder == "" {
			log.Fatal("No output folder specified")
		}

		target = appdeploy.NewFolderTarget(folder)
	*/

	target = appdeploy.NewKubernetesTarget("vagrant-single")

	err := appdeploy.Process(src, target)
	if err != nil {
		log.Fatal(err)
	}
}
