package command

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/trojan295/k8s-utils/pkg/k8siperf3"
)

var (
	serverNodeName string
	clientNodeName string
	parallelCount  int
	kubeconfig     string
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8siperf3",
		Short: "k8siperf3 is a tool to run iperf3 tests on Kubernetes",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientset, err := k8siperf3.GetClientset(kubeconfig)
			if err != nil {
				return fmt.Errorf("while getting clientset: %w", err)
			}

			iperf3 := k8siperf3.NewK8sIperf3(clientset)

			result, err := iperf3.Run(cmd.Context(), &k8siperf3.Config{
				ServerNodeName: serverNodeName,
				ClientNodeName: clientNodeName,
				ParallelCount:  parallelCount,
			})
			if err != nil {
				return fmt.Errorf("while running iperf3: %w", err)
			}

			fmt.Printf("Bitrate: %f Gbits/sec\n", result.Bitrate)

			return nil
		},
	}

	var kubeconfigDefault string
	if os.Getenv("KUBECONFIG") != "" {
		kubeconfigDefault = os.Getenv("KUBECONFIG")
	} else {
		home, _ := os.UserHomeDir()
		if home != "" {
			kubeconfigDefault = path.Join(home, ".kube", "config")
		} else {
			kubeconfigDefault = ".kubeconfig"
		}
	}

	cmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", kubeconfigDefault, "Absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringVar(&serverNodeName, "server-node-name", "", "Node name for the iperf3 server pods")
	cmd.PersistentFlags().StringVar(&clientNodeName, "client-node-name", "", "Node name for the iperf3 client pods")
	cmd.PersistentFlags().IntVar(&parallelCount, "parallel-count", 1, "Number of parallel iperf3 tests")

	return cmd
}
