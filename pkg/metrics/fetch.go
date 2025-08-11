package fetch

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// UsageStats holds CPU & Memory averages
type UsageStats struct {
	AvgCPU    float64 // in millicores
	AvgMemory float64 // in MiB
}

// getKubeClients returns both core and metrics clients
func getKubeClients() (*kubernetes.Clientset, *metricsv.Clientset, error) {
	// Try in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	// Register for metrics API
	if err := metricsv.AddToScheme(scheme.Scheme); err != nil {
		log.Printf("warning: failed to add metrics scheme: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create core client: %w", err)
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return clientset, metricsClient, nil
}

// getPodMetrics fetches CPU (millicores) and Memory (MiB) for all pods matching a selector
func getPodMetrics(namespace, selector string) ([]UsageStats, error) {
	_, metricsClient, err := getKubeClients()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	podMetricsList, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	var results []UsageStats
	for _, m := range podMetricsList.Items {
		var totalCPU, totalMem float64
		for _, c := range m.Containers {
			cpuQty := c.Usage.Cpu().MilliValue() // millicores
			memQty := c.Usage.Memory().Value()   // bytes
			totalCPU += float64(cpuQty)
			totalMem += float64(memQty) / (1024 * 1024) // MiB
		}
		results = append(results, UsageStats{
			AvgCPU:    totalCPU,
			AvgMemory: totalMem,
		})
	}

	return results, nil
}

// GetDeploymentAverageUsage calculates avg CPU & memory for all pods in a deployment
func GetDeploymentAverageUsage(namespace, deploymentName string) (UsageStats, error) {
	clientset, _, err := getKubeClients()
	if err != nil {
		return UsageStats{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dep, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return UsageStats{}, fmt.Errorf("failed to get deployment: %w", err)
	}

	selector := metav1.FormatLabelSelector(dep.Spec.Selector)

	podStats, err := getPodMetrics(namespace, selector)
	if err != nil {
		return UsageStats{}, err
	}

	if len(podStats) == 0 {
		return UsageStats{}, fmt.Errorf("no pods found for deployment %s", deploymentName)
	}

	var totalCPU, totalMem float64
	for _, s := range podStats {
		totalCPU += s.AvgCPU
		totalMem += s.AvgMemory
	}

	return UsageStats{
		AvgCPU:    totalCPU / float64(len(podStats)),
		AvgMemory: totalMem / float64(len(podStats)),
	}, nil
}
