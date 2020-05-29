package iperf

import (
	"context"
	"fmt"
	"time"

	iperfv1alpha1 "github.com/jharrington22/iperf-operator/pkg/apis/iperf/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

const (
	requeueWaitTime = time.Duration(1)
)

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
	cr := &iperfv1alpha1.Iperf{}
	err := r.client.Get(ctx, request.NamespacedName, cr)
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

	// Fetch configuration for iPerf client/server's
	sessionDuration := cr.Spec.SessionDuration
	concurrentConnections := cr.Spec.ConcurrentConnections

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

	iperfServers := make(map[string]string)

	for _, label := range workerNodeLabels {
		serverNamePrefix := "iperf-server-"
		namespacedName := types.NamespacedName{
			Name:      fmt.Sprintf("%s%s", serverNamePrefix, label),
			Namespace: request.Namespace,
		}
		// Create server pod on worker node
		iperfServerPod := generateServerPod(namespacedName, label)

		// Set Iperf instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, iperfServerPod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if a server pod already exists
		found := &corev1.Pod{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: namespacedName.Name, Namespace: namespacedName.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new iperf server Pod", "iperfServerPod.Namespace", iperfServerPod.Namespace, "iperfServerPod.Name", iperfServerPod.Name, "iperServerPodWorkerLabel", label)
			err = r.client.Create(context.TODO(), iperfServerPod)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}
		// Continue as server pod alraedy exists

		// Get server pod IP to pass to iPerf clients
		iperfServerIP, err := r.getServerPodIP(namespacedName)
		if err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, object my not be ready yet
				// TODO: Wait for object to exist instead of requeuing in time
				// Log that the object was not found and that we are requeuing (Assuming it'll exist in the future)
				reqLogger.Info(fmt.Sprintf("Unable to find pod %s retrying in %d", namespacedName.Name, requeueWaitTime))
				return reconcile.Result{RequeueAfter: requeueWaitTime}, nil
			}
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}

		iperfServers[label] = *iperfServerIP
	}

	for label, iperfServerIP := range iperfServers {
		clientNamePrefix := "iperf-client-"
		namespacedName := types.NamespacedName{
			Name:      fmt.Sprintf("%s%s", clientNamePrefix, label),
			Namespace: request.Namespace,
		}
		iperfClientPod := generateClientPod(namespacedName, iperfServerIP, label, sessionDuration, concurrentConnections)

		// Set Iperf instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, iperfClientPod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if a server pod already exists
		found := &corev1.Pod{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: namespacedName.Name, Namespace: namespacedName.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new iperf server Pod", "iperfClientPod.Namespace", iperfClientPod.Namespace, "iperfClientPod.Name", iperfClientPod.Name, "iperClientPodWorkerLabel", label)
			err = r.client.Create(context.TODO(), iperfClientPod)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}
		// Continue as client pod alraedy exists
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileIperf) getServerPodIP(namespacedName types.NamespacedName) (*string, error) {

	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), namespacedName, pod)
	if err != nil {
		// Return the error to the recoile so it can requeue correctly
		return nil, err
	}

	// Return pod IP
	return &pod.Status.PodIP, nil
}
