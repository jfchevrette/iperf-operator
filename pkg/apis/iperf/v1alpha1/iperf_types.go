package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IperfSpec defines the desired state of Iperf
type IperfSpec struct {
	// MaxThroughput Max bandwidth all clients should consume (divided equally between num of clients)
	MaxThroughput int `json:"maxThroughput,omitempty"`
	// ConcurrentConnections Total number of connections from client to server (divided equally between num of clients)
	ConcurrentConnections int `json:"concurrentConnections"`
	// SessionDuration duration in minutes for the test to run
	SessionDuration int `json:"sessionDuration"`
	// ClientNum Number of clients, should not exceed number of nodes (default == number of nodes)
	ClientNum int `json:"clientNum,omitempty"`
	// ServerNum Number of servers, should not exceed number of nodes (default 1)
	ServerNum int `json:"serverNum,omitempty"`
}

// IperfStatus defines the observed state of Iperf
type IperfStatus struct {
	// TODO add conntrack max and number of connections status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Iperf is the Schema for the iperves API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=iperves,scope=Namespaced
type Iperf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IperfSpec   `json:"spec,omitempty"`
	Status IperfStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IperfList contains a list of Iperf
type IperfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Iperf `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Iperf{}, &IperfList{})
}
