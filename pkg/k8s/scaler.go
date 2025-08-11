package k8s

import (
	"context"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ScaleDeployment(namespace, name string, replicas int32){
	client := GetClient()

	dep, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error getting deployment: %v",err)
	}

	dep.Spec.Replicas := &replicas
	_, err := client.AppsV1().Deployments(namespace).Update(context.TODO(), dep, metav1.UpdateOptions{})

	if err != nil {
		log.Fatalf("Error scaling deployment: %v", err)
	}

	log.Fatalf("Scaled %s to %d replicas ", name, replicas)
	
}