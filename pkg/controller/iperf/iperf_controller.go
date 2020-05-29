package iperf

import (
	"context"
	"fmt"

	iperfv1alpha1 "github.com/jharrington22/iperf-operator/pkg/apis/iperf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_iperf")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Iperf Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIperf{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("iperf-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Iperf
	err = c.Watch(&source.Kind{Type: &iperfv1alpha1.Iperf{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Iperf
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &iperfv1alpha1.Iperf{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileIperf implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIperf{}

// ReconcileIperf reconciles a Iperf object
type ReconcileIperf struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Iperf object and makes changes based on the state read
// and what is in the Iperf.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIperf) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Iperf")

	// Set context
	ctx := context.TODO()

	// Fetch the Iperf instance
	instance := &iperfv1alpha1.Iperf{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Fetch a list of worker nodes on the cluster
	workerNodeList := &corev1.NodeList{}
	workerNodeListOpts := []client.ListOption{
		client.MatchingLabels{
			nodeWorkerSelectorKey: nodeWorkerSelectorValue,
		},
	}
	err = r.client.List(ctx, workerNodeList, workerNodeListOpts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Determine the number of worker nodes
	workerNodeNum := len(workerNodeList.Items)

	reqLogger.Info(fmt.Sprintf("%d worker nodes found", workerNodeNum))

	workerNodeLabels := getWorkerNodeLabels(workerNodeList)

	workerNodes := r.discorverWorkerNodes(workerNodeList)

	for _, label := range workerNodeLabels {
		createClientPod()
	}

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set Iperf instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

func (r *ReconcileIperf) discorverWorkerNodes() map[string]stirng {

	workerNodeMeta := map[string]stirng

	for _, workerNode := range workerNodeList.Items {
		label := getWorkerNodeLabel(workerNode)
		ip := getWorkerNodeIP(workerNode)

		workerNodeMeta[label]{ip}
	}

	return workerNodeMeta
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *iperfv1alpha1.Iperf) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func getWorkerNodeLabel(workerNode *corev1.Node) string {
	return workerNode.Labels["kubernetes.io/hostname"]
}

// getWorkerNodeIP returns the internal IP of the  
func getWrokerNodeIP(workerNode *corev1.Node) *string {
	var internalIP string

	for key, value := range workerNode.Status.Addresses {
		if key == nodeInternalIPKey {
			internalIP = value 
		}
	}
	return &internalIP

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