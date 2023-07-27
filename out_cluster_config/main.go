package main

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main() {
	homePath := homedir.HomeDir()

	kubeConfig := filepath.Join(homePath, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)

	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		pods, err := clientSet.CoreV1().Pods("default").List(
			context.TODO(),
			metav1.ListOptions{},
		)
		if err != nil {
			log.Fatal(err)
		}
		for index, pod := range pods.Items {
			log.Printf("%d: %s -> %s\n", index+1, pod.Namespace, pod.GetName())
		}
		<-time.Tick(5 * time.Second)
	}
}
