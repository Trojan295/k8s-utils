# k8siperf3

k8siperf3 is a tool to run iperf3 tests on Kubernetes

```
Usage:
  k8siperf3 [flags]

Flags:
      --client-node-name string   Node name for the iperf3 client pods
      --cooldown-duration int     Duration of the iperf3 cooldown in seconds (default 15)
  -h, --help                      help for k8siperf3
      --kubeconfig string         Absolute path to the kubeconfig file (default "$HOME/.kube/config")
      --parallel-count int        Number of parallel iperf3 tests (default 1)
      --server-node-name string   Node name for the iperf3 server pods
      --test-duration int         Duration of the iperf3 test in seconds (default 30)
      --warmup-duration int       Duration of the iperf3 warmup in seconds (default 15)
```
