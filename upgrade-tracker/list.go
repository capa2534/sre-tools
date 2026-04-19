package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func listUpgrades() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	gvr := schema.GroupVersionResource{
		Group:    "sre.io",
		Version:  "v1",
		Resource: "clientupgrades",
	}

	list, err := dynClient.Resource(gvr).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listando upgrades: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%-20s %-15s %-15s\n", "CLIENT", "VERSION", "STATUS")
	fmt.Println("----------------------------------------------")

	for _, item := range list.Items {
		spec := item.Object["spec"].(map[string]interface{})
		clientName := spec["clientName"].(string)
		version := spec["targetVersion"].(string)
		status := spec["status"].(string)

		icon := "⏳"
		switch status {
		case "completed":
			icon = "✅"
		case "in-progress":
			icon = "🔄"
		case "failed":
			icon = "❌"
		}

		fmt.Printf("%-20s %-15s %s %s\n", clientName, version, icon, status)
	}
}
