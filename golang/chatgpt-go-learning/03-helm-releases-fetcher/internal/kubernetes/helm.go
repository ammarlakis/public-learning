package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetHelmReleaseResources fetches all resources from a specific Helm release and their statuses
func GetHelmReleaseResources(releaseName string) ([]unstructured.Unstructured, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)

	// Initialize Helm configuration
	if err := actionConfig.Init(settings.RESTClientGetter(), "", "secrets", log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action config: %w", err)
	}

	// Fetch the Helm release
	getAction := action.NewGet(actionConfig)
	release, err := getAction.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Helm release %s: %w", releaseName, err)
	}

	releaseNamespace := release.Namespace
	manifest := release.Manifest
	if manifest == "" {
		return nil, fmt.Errorf("no manifest found for release %s", releaseName)
	}

	// Decode YAML manifest into Kubernetes objects
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	resources := []unstructured.Unstructured{}
	yamlDocs := bytes.Split([]byte(manifest), []byte("\n---\n"))

	for _, doc := range yamlDocs {
		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		var obj unstructured.Unstructured
		_, _, err := decoder.Decode(doc, nil, &obj)
		if err != nil {
			log.Printf("Skipping invalid YAML document: %v", err)
			continue
		}

		if obj.GetNamespace() == "" {
			obj.SetNamespace(releaseNamespace)
		}

		resources = append(resources, obj)
	}

	// Fetch live statuses and events
	err = fetchResourceStatuses(resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// fetchResourceStatuses retrieves the status and last 10 events of each resource
func fetchResourceStatuses(resources []unstructured.Unstructured) error {
	restConfig, err := getKubeConfig()
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes dynamic client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create discovery client: %w", err)
	}

	for i, resource := range resources {
		kind, namespace, name := resource.GetKind(), resource.GetNamespace(), resource.GetName()

		gvr, err := getGVR(discoveryClient, kind)
		if err != nil {
			log.Printf("Skipping resource %s/%s: %v", namespace, name, err)
			continue
		}

		obj, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			log.Printf("Failed to get live resource %s/%s: %v", namespace, name, err)
			continue
		}

		status, found, _ := unstructured.NestedMap(obj.Object, "status")
		if found {
			resources[i].Object["status"] = status
		} else {
			resources[i].Object["status"] = map[string]interface{}{"message": "No status available"}
		}

		events, err := fetchLast10Events(clientset, namespace, name, kind)
		if err != nil {
			log.Printf("Failed to fetch events for %s/%s: %v", namespace, name, err)
		}
		resources[i].Object["events"] = events
	}

	return nil
}

// fetchLast10Events retrieves the last 10 events related to a resource
func fetchLast10Events(clientset *kubernetes.Clientset, namespace, name, kind string) ([]map[string]interface{}, error) {
	eventList, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var filteredEvents []map[string]interface{}
	for _, event := range eventList.Items {
		if event.InvolvedObject.Name == name && event.InvolvedObject.Kind == kind {
			filteredEvents = append(filteredEvents, map[string]interface{}{
				"reason":  event.Reason,
				"message": event.Message,
				"type":    event.Type,
				"count":   event.Count,
				"time":    event.LastTimestamp,
			})
		}
	}

	sort.Slice(filteredEvents, func(i, j int) bool {
		timeI, _ := filteredEvents[i]["time"].(metav1.Time) // Extract metav1.Time
		timeJ, _ := filteredEvents[j]["time"].(metav1.Time) // Extract metav1.Time
		return timeI.Time.After(timeJ.Time)                 // Compare time.Time values
	})

	if len(filteredEvents) > 10 {
		filteredEvents = filteredEvents[:10]
	}

	return filteredEvents, nil
}

// getKubeConfig loads Kubernetes config dynamically
func getKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		log.Println("Using in-cluster Kubernetes config")
		return config, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find home directory: %w", err)
	}

	kubeconfigPath := filepath.Join(home, ".kube", "config")
	log.Printf("Using local kubeconfig: %s", kubeconfigPath)

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config, nil
}

// getGVR dynamically maps a resource kind to a GroupVersionResource (GVR)
func getGVR(discoveryClient *discovery.DiscoveryClient, kind string) (schema.GroupVersionResource, error) {
	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to list API resources: %w", err)
	}

	for _, apiResourceList := range apiResourceLists {
		for _, apiResource := range apiResourceList.APIResources {
			if apiResource.Kind == kind {
				groupVersion, _ := schema.ParseGroupVersion(apiResourceList.GroupVersion)
				return schema.GroupVersionResource{Group: groupVersion.Group, Version: groupVersion.Version, Resource: apiResource.Name}, nil
			}
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("resource kind %s not found", kind)
}
