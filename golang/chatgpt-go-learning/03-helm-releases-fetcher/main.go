package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ammarlakis/helm-releases-fetcher/internal/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./helm-releases-fetcher <release-name>")
		os.Exit(1)
	}

	releaseName := os.Args[1]
	fmt.Printf("Fetching resources for Helm release: %s\n", releaseName)

	// Fetch resources
	resources, err := kubernetes.GetHelmReleaseResources(releaseName)
	if err != nil {
		log.Fatalf("Error fetching resources: %v", err)
	}

	// Print resources with statuses and events
	if len(resources) == 0 {
		fmt.Println("No resources found.")
	} else {
		fmt.Println("Resources:")
		for _, resource := range resources {
			fmt.Printf("- Kind: %s | Name: %s | Namespace: %s\n",
				resource.GetKind(), resource.GetName(), resource.GetNamespace())

			// Print status if available
			printResourceStatus(resource)

			// Print last 10 events
			printResourceEvents(resource)
		}
	}
}

// printResourceStatus prints the status of a Kubernetes resource
func printResourceStatus(resource unstructured.Unstructured) {
	status, found, _ := unstructured.NestedMap(resource.Object, "status")
	if found {
		fmt.Printf("  Status: %v\n", status)
	} else {
		fmt.Println("  Status: No status available")
	}
}

// printResourceEvents prints the last 10 events for a resource
func printResourceEvents(resource unstructured.Unstructured) {
	events, found, _ := unstructured.NestedSlice(resource.Object, "events")
	if !found || len(events) == 0 {
		fmt.Println("  Events: No events found")
		return
	}

	fmt.Println("  Events:")
	for _, event := range events {
		eventMap := event.(map[string]interface{})
		fmt.Printf("    - Reason: %s | Message: %s | Type: %s | Count: %v | Time: %v\n",
			eventMap["reason"], eventMap["message"], eventMap["type"], eventMap["count"], eventMap["time"])
	}
}
