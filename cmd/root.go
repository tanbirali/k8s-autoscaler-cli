package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tanbirali/k8s-autoscaler-cli/pkg/k8s"
)


var (
	namespace 			string
	deployment			string
	cpuThreshold		int
	memoryThreshold 	int
	interval			time.Duration
	dryRun				bool
)



var rootCmd = &cobra.Command{
	Use: "K8s-autoscaler",
	Short: "Auto-scales Kubernetes deployments based on CPU",
	Run: func(cmd *cobra.Command, args []string) {
		if deployment == "" {
			log.Fatal("You must provide a deployment name with --deployment")
		}

		fmt.Printf("🚀 Starting Auto-Scaler for %s in namespace %s\n", deployment, namespace)
		client := k8s.GetClient()

		for {
			avgCPU, avgMem, err := metrics.GetDeploymentAverageUsage(namespace, deployment)
			if err != nil {
				log.Printf("Error getting metrics: %v", err)
				time.Sleep(interval)
				continue
			}

			fmt.Printf("📊 Avg CPU: %.2f%%, Avg Memory: %.2f%%\n", avgCPU, avgMem)

			dep, _ := client.AppsV1().Deployments(namespace).Get(context.TODO(), deployment, metaV1.GetOptions{})
			currentReplicas := *dep.Spec.Replicas
			newReplicas := currentReplicas

			// Scale Up
			if avgCPU > float64(cpuThreshold) || avgMem > float64(memoryThreshold) {
				newReplicas = currentReplicas + 1
			}

			// Scale Down (minimum 1 replica)
			if avgCPU < float64(cpuThreshold-10) && avgMem < float64(memoryThreshold-10) && currentReplicas > 1 {
				newReplicas = currentReplicas - 1
			}

			if newReplicas != currentReplicas {
				if dryRun {
					fmt.Printf("💡 Dry-run: Would scale %s from %d → %d replicas\n", deployment, currentReplicas, newReplicas)
				} else {
					k8s.ScaleDeployment(namespace, deployment, newReplicas)
				}
			} else {
				fmt.Println("✅ No scaling action needed")
			}

			time.Sleep(interval)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringVarP(&deployment, "deployment", "d", "", "Deployment name")
    rootCmd.PersistentFlags().IntVar(&cpuThreshold, "cpu-threshold", 70, "CPU usage % threshold")
    rootCmd.PersistentFlags().IntVar(&memoryThreshold, "memory-threshold", 80, "Memory usage % threshold")
    rootCmd.PersistentFlags().DurationVar(&interval, "interval", 30*time.Second, "Check interval")
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Run without scaling")
}
func Execute(){
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}