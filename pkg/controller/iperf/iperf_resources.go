package iperf

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	iperfServerImage        = "quay.io/jharrington22/network-toolbox:latest"
	iperfClientImage        = "quay.io/jharrington22/network-toolbox:latest"
	iperfCmd                = "iperf"
	nodeSelectorKey         = "kubernetes.io/hostname"
	nodeWorkerSelectorKey   = "node-role.kubernetes.io/worker"
	nodeWorkerSelectorValue = ""
)

func generateServerPod(namespacedName types.NamespacedName, nodeSelectorValue string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "iperf-server",
					Image:   iperfServerImage,
					Command: []string{iperfCmd},
					Args:    []string{"-s"},
				},
			},
			NodeSelector: map[string]string{
				nodeSelectorKey: nodeSelectorValue,
			},
		},
	}

}

func generateClientPod(namespacedName types.NamespacedName, podIP, nodeSelectorValue string, sessionDuration, concurrentConnections int) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "iperf-client",
					Image:   iperfClientImage,
					Command: []string{iperfCmd},
					Args:    []string{"-c", podIP, "-t", string(sessionDuration), "-P", string(concurrentConnections)},
				},
			},
			NodeSelector: map[string]string{
				nodeSelectorKey: nodeSelectorValue,
			},
		},
	}

}
