package iperf

import (
	corev1 "k8s.io/api/core/v1"
)

func getWorkerNodeLabel(workerNode *corev1.Node) string {
	return workerNode.Labels["kubernetes.io/hostname"]
}

func getWorkerNodeLabels(workerNodeList *corev1.NodeList) []string {
	// Populate a list of nodes to generate server/client pods for by label
	// The label we'll use is kubernetes.io/hostname=ip-10-100-138-234
	workerNodeLabels := []string{}

	for _, workerNode := range workerNodeList.Items {
		workerNodeLabels = append(workerNodeLabels, workerNode.Labels["kubernetes.io/hostname"])
	}

	return workerNodeLabels
}
