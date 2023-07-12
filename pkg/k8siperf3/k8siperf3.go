package k8siperf3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func GetClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("while building config from flags: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("while creating clientset: %w", err)
	}

	return clientset, nil
}

type K8sIperf3 struct {
	clientset *kubernetes.Clientset
}

type Config struct {
	ServerNodeName string
	ClientNodeName string
	ParallelCount  int
}

type Result struct {
	Bitrate float64
}

func aggregateResults(results ...*Result) *Result {
	result := &Result{
		Bitrate: 0,
	}

	for _, r := range results {
		result.Bitrate += r.Bitrate
	}

	return result
}

func NewK8sIperf3(clientset *kubernetes.Clientset) *K8sIperf3 {
	return &K8sIperf3{
		clientset: clientset,
	}
}

func (iperf *K8sIperf3) Run(ctx context.Context, cfg *Config) (*Result, error) {
	serverIPs := make(map[int]string)

	log.Printf("Creating %d iperf3 server pods\n", cfg.ParallelCount)

	for i := 0; i < cfg.ParallelCount; i++ {
		if err := iperf.createServer(ctx, cfg.ServerNodeName, i); err != nil {
			return nil, fmt.Errorf("while creating server: %w", err)
		}

		defer func(i int) {
			if err := deleteServer(ctx, iperf.clientset, i); err != nil {
				log.Printf("failed to delete server %d: %v\n", i, err)
			}
		}(i)
	}

	log.Printf("Waiting for iperf3 server pods to start\n")

	for i := 0; i < cfg.ParallelCount; i++ {
		for {
			ip, err := iperf.getServerIPAddress(ctx, i)
			if err != nil {
				return nil, fmt.Errorf("while getting server IP address: %w", err)
			}

			if ip != "" {
				serverIPs[i] = ip
				log.Printf("iperf3 server pod %d IP address: %s\n", i, ip)

				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("Creating %d iperf3 client pods\n", cfg.ParallelCount)

	for i := 0; i < cfg.ParallelCount; i++ {
		if err := iperf.createClient(ctx, cfg.ClientNodeName, i, serverIPs[i]); err != nil {
			return nil, fmt.Errorf("while creating client: %w", err)
		}

		defer func(i int) {
			if err := deleteClient(ctx, iperf.clientset, i); err != nil {
				log.Printf("failed to delete client %d: %v\n", i, err)
			}
		}(i)
	}

	log.Printf("Waiting for iperf3 client pods to finish\n")

	results := make([]*Result, cfg.ParallelCount)

	for i := 0; i < cfg.ParallelCount; i++ {
		for {
			pod, err := iperf.clientset.CoreV1().Pods("default").Get(ctx, fmt.Sprintf("iperf3-client-%d", i), metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("while getting client pod: %w", err)
			}

			if pod.Status.Phase == v1.PodSucceeded {
				log.Printf("iperf3 client pod %d finished\n", i)

				logs, err := iperf.getClientLogs(ctx, i)
				if err != nil {
					return nil, fmt.Errorf("while getting client logs: %w", err)
				}

				result, err := parseIperf3Logs(logs, 10, 10)
				if err != nil {
					return nil, fmt.Errorf("while parsing client logs: %w", err)
				}

				results[i] = result

				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	return aggregateResults(results...), nil
}

func (iperf *K8sIperf3) createServer(ctx context.Context, nodeName string, count int) error {
	name := fmt.Sprintf("iperf3-server-%d", count)

	if _, err := iperf.clientset.CoreV1().Pods("default").Create(ctx, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
			Containers: []v1.Container{{
				Name:  "iperf3",
				Image: "networkstatic/iperf3",
				Args:  []string{"-s"},
			}},
			RestartPolicy: v1.RestartPolicyAlways,
		},
	}, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (iperf *K8sIperf3) createClient(ctx context.Context, nodeName string, count int, ip string) error {
	name := fmt.Sprintf("iperf3-client-%d", count)

	if _, err := iperf.clientset.CoreV1().Pods("default").Create(ctx, &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
			Containers: []v1.Container{{
				Name:  "iperf3",
				Image: "networkstatic/iperf3",
				Args: []string{
					"-f", "g",
					"-t", "70",
					"-c", ip,
				},
			}},
			RestartPolicy: v1.RestartPolicyOnFailure,
		}}, metav1.CreateOptions{}); err != nil {
		return err
	}

	return nil
}

func (iperf *K8sIperf3) getServerIPAddress(ctx context.Context, count int) (string, error) {
	name := fmt.Sprintf("iperf3-server-%d", count)
	pod, err := iperf.clientset.CoreV1().Pods("default").Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return pod.Status.PodIP, nil
}

func (iperf *K8sIperf3) getClientLogs(ctx context.Context, count int) (string, error) {
	name := fmt.Sprintf("iperf3-client-%d", count)

	podLogOpts := v1.PodLogOptions{}
	req := iperf.clientset.CoreV1().Pods("default").GetLogs(name, &podLogOpts)

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("while getting client logs: %w", err)
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, podLogs); err != nil {
		return "", fmt.Errorf("while copying client logs: %w", err)
	}

	return buf.String(), nil
}

func deleteClient(ctx context.Context, clientset *kubernetes.Clientset, count int) error {
	return deletePod(ctx, clientset, fmt.Sprintf("iperf3-client-%d", count))
}

func deleteServer(ctx context.Context, clientset *kubernetes.Clientset, count int) error {
	return deletePod(ctx, clientset, fmt.Sprintf("iperf3-server-%d", count))
}

func deletePod(ctx context.Context, clientset *kubernetes.Clientset, name string) error {
	return clientset.CoreV1().Pods("default").Delete(ctx, name, metav1.DeleteOptions{})
}
