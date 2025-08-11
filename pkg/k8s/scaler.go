package k8s

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ScaleDeployment scales a Kubernetes deployment to the desired number of replicas.
func ScaleDeployment(namespace, name string, replicas int32) error {
	client := GetClient()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the deployment
	dep, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error getting deployment %q: %w", name, err)
	}

	// Set replicas
	dep.Spec.Replicas = &replicas

	// Update the deployment
	_, err = client.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error scaling deployment %q: %w", name, err)
	}

	log.Printf("✅ Scaled %s to %d replicas", name, replicas)
	return nil
}
