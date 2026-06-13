package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Shashank0701-byte/Loadster/pkg/config"
	"github.com/Shashank0701-byte/Loadster/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	namespace  string
	deployment string
)

var distributedCmd = &cobra.Command{
	Use:   "distributed",
	Short: "Run load tests in distributed mode",
	Long:  `Orchestrates load test execution in distributed mode using Kubernetes.`,
}

var distributedRunCmd = &cobra.Command{
	Use:   "run [scenario.yaml]",
	Short: "Run a distributed load test",
	Long:  `Deploys workers and coordinates a scenario-based load test across a Kubernetes cluster.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("scenario.yaml file is required")
		}

		scenarioFile := args[0]
		logger.Log.Info("Loading test scenario from YAML (Distributed Mode)", zap.String("file", scenarioFile))
		cfg, err := config.ParseFile(scenarioFile)
		if err != nil {
			return fmt.Errorf("failed to load scenario: %w", err)
		}

		logger.Log.Info("Connecting to Kubernetes cluster...")
		k8sClient, err := getK8sClient()
		if err != nil {
			return fmt.Errorf("failed to connect to Kubernetes: %w", err)
		}
		logger.Log.Info("Connected to Kubernetes successfully")

		ctx := cmd.Context()
		for i, stage := range cfg.Stages {
			logger.Log.Info("Distributed Execution Stage",
				zap.Int("stage_number", i+1),
				zap.Int("workers_target", stage.Users),
				zap.String("duration", stage.RawDuration),
			)

			logger.Log.Info("Scaling worker deployment...", zap.String("deployment", deployment), zap.Int("replicas", stage.Users))
			err = scaleDeployment(ctx, k8sClient, namespace, deployment, int32(stage.Users))
			if err != nil {
				return fmt.Errorf("failed to scale worker deployment: %w", err)
			}

			select {
			case <-ctx.Done():
				logger.Log.Info("Distributed execution cancelled, scaling down workers...")
				_ = scaleDeployment(context.Background(), k8sClient, namespace, deployment, 0)
				return ctx.Err()
			case <-time.After(stage.Duration):
				logger.Log.Info("Stage completed", zap.Int("stage_number", i+1))
			}
		}

		logger.Log.Info("Test execution finished. Scaling down workers...")
		err = scaleDeployment(context.Background(), k8sClient, namespace, deployment, 0)
		if err != nil {
			logger.Log.Error("Failed to scale down worker deployment", zap.Error(err))
		}

		return nil
	},
}

func getK8sClient() (*kubernetes.Clientset, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return kubernetes.NewForConfig(config)
	}

	// Fallback to local kubeconfig
	var kubeconfigPath string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load kubeconfig: %w", err)
	}

	return kubernetes.NewForConfig(config)
}

func scaleDeployment(ctx context.Context, client *kubernetes.Clientset, ns, name string, replicas int32) error {
	scale, err := client.AppsV1().Deployments(ns).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = client.AppsV1().Deployments(ns).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	return err
}

func init() {
	distributedRunCmd.Flags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	distributedRunCmd.Flags().StringVar(&deployment, "deployment", "loadster-worker", "Kubernetes worker deployment name")

	distributedCmd.AddCommand(distributedRunCmd)
	rootCmd.AddCommand(distributedCmd)
}
type DistributedCmd_Type = *cobra.Command // For linking symbol reference
