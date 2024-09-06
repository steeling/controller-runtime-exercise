/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"reflect"
	"time"

	"github.com/go-logr/logr"
	myappv1 "github.com/steeling/controller-runtime-exercise/api/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	policyv1 "k8s.io/api/policy/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/prometheus/client_golang/prometheus"
	//	"github.com/steeling/controller-runtime-exercise/pkg/api"
	//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	//	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	reconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myapp_reconcile_duration_seconds",
			Help:    "Duration of the MyApp reconciliation loop",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)
	myAppInstances = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "myapp_instances",
			Help: "Number of MyApp instances currently managed by the controller",
		},
		[]string{"namespace"},
	)
	reconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_reconcile_errors_total",
			Help: "Total number of errors encountered during reconciliation",
		},
		[]string{"kind", "message"},
	)
)

func init() {
	// Register custom metrics with the global Prometheus registry
	metrics.Registry.MustRegister(reconcileDuration, myAppInstances, reconcileErrors)
}

// MyAppReconciler reconciles a MyApp object
type MyAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=myapp.example.com,resources=myapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=myapp.example.com,resources=myapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=myapp.example.com,resources=myapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("myapp", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime).Seconds()
		reconcileDuration.WithLabelValues("success").Observe(duration)
	}()

	//Fetch the MyApp instance
	var myApp myappv1.MyApp
	if err := r.Client.Get(ctx, req.NamespacedName, &myApp); err != nil {
		if apierrors.IsNotFound(err) {
			// MyApp resource not found. Could have been deleted
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		log.Error(err, "Failed to get MyApp")
		reconcileErrors.WithLabelValues("GetMyApp", err.Error()).Inc()
		reconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}
	// Check if MyApp is marked for deletion
	if myApp.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add finalizer if it's not already present
		if !containsString(myApp.ObjectMeta.Finalizers, "myapp.finalizers.example.com") {
			myApp.ObjectMeta.Finalizers = append(myApp.ObjectMeta.Finalizers, "myapp.finalizers.example.com")
			if err := r.Client.Update(ctx, &myApp); err != nil {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(myApp.ObjectMeta.Finalizers, "myapp.finalizers.example.com") {
			// Handle deletion of child resources
			if err := r.cleanupResources(ctx, &myApp); err != nil {
				log.Error(err, "Failed to cleanup child resources")
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}

			// Remove finalizer after cleanup
			myApp.ObjectMeta.Finalizers = removeString(myApp.ObjectMeta.Finalizers, "myapp.finalizers.example.com")
			if err := r.Client.Update(ctx, &myApp); err != nil {
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}
		}
		// Stop reconciliation as the object is being deleted
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
	myAppInstances.WithLabelValues(myApp.Namespace).Inc()
	defer myAppInstances.WithLabelValues(myApp.Namespace).Dec()

	// Check if the MyApp instance has the label "reconciler: ignore"
	if val, exists := myApp.Labels["reconciler"]; exists && val == "ignore" {
		log.Info("Ignoring MyApp resource due to reconciler: ignore label", "namespace", myApp.Namespace, "name", myApp.Name)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// desired Deployment based on MyAppSpec with default resources
	desiredDeployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myApp.Name + "-deployment",
			Namespace: myApp.Namespace,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: myApp.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": myApp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": myApp.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "myapp-container",
							Image: myApp.Spec.Image,
							Args:  myApp.Spec.Args,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 80},
							},
						},
					},
				},
			},
		},
	}

	// PodDisruptionBudget:
	desiredPDB := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      myApp.Name + "-pdb",
			Namespace: myApp.Namespace,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &intstr.IntOrString{IntVal: 1},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": myApp.Name},
			},
		},
	}

	// If dep doesn't exist create it and log errors
	var existingDeployment appv1.Deployment
	err := r.Client.Get(ctx, types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, &existingDeployment)
	if err != nil && apierrors.IsNotFound(err) {
		// Create new dep
		log.Info("Creating Deployment", "namespace", desiredDeployment.Namespace, "name", desiredDeployment.Name)
		if err := r.Client.Create(ctx, desiredDeployment); err != nil {
			log.Error(err, "Failed to create Deployment", "namespace", desiredDeployment.Namespace, "name", desiredDeployment.Name)
			reconcileErrors.WithLabelValues("CreateDeployment", err.Error()).Inc()
			reconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		reconcileErrors.WithLabelValues("GetDeployment", err.Error()).Inc()
		reconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	} else {
		// Update the Deployment if it differs from the desired state
		if !reflect.DeepEqual(existingDeployment.Spec, desiredDeployment.Spec) {
			log.Info("Updating Deployment", "namespace", existingDeployment.Namespace, "name", existingDeployment.Name)
			existingDeployment.Spec = desiredDeployment.Spec
			if err := r.Client.Update(ctx, &existingDeployment); err != nil {
				log.Error(err, "Failed to update Deployment", "namespace", existingDeployment.Namespace, "name", existingDeployment.Name)
				reconcileErrors.WithLabelValues("UpdateDeployment", err.Error()).Inc()
				reconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
				return ctrl.Result{RequeueAfter: 30 * time.Second}, err
			}
		}
	}

	// If PodDisruptionBudget doesn't exist create it and log errors
	var existingPDB policyv1.PodDisruptionBudget
	err = r.Client.Get(ctx, types.NamespacedName{Name: desiredPDB.Name, Namespace: desiredPDB.Namespace}, &existingPDB)
	if err != nil && apierrors.IsNotFound(err) {
		// Create new PDB
		log.Info("Creating PodDisruptionBudget", "namespace", desiredPDB.Namespace, "name", desiredPDB.Name)
		if err := r.Client.Create(ctx, desiredPDB); err != nil {
			log.Error(err, "Failed to create PodDisruptionBudget", "namespace", desiredPDB.Namespace, "name", desiredPDB.Name)
			reconcileErrors.WithLabelValues("CreatePDB", err.Error()).Inc()
			reconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get PodDisruptionBudget")
		reconcileErrors.WithLabelValues("GetPDB", err.Error()).Inc()
		reconcileDuration.WithLabelValues("error").Observe(time.Since(startTime).Seconds())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}
	//Make Sure MyApp is up to date
	if myApp.Status.Phase != "Running" || !myApp.Status.Healthy {
		myApp.Status.Phase = "Running"
		myApp.Status.Healthy = true
		if err := r.Client.Status().Update(ctx, &myApp); err != nil {
			log.Error(err, "Failed to update MyApp status")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err
		}
	}
	/*	if err := ctrl.SetControllerReference(&myApp, desiredDeployment, c.manager.GetScheme()); err != nil {
			return ctrl.Result{}, err
		}
		if err := ctrl.SetControllerReference(&myApp, desiredPDB, c.manager.GetScheme()); err != nil {
			return ctrl.Result{}, err
		}*/

	reconcileDuration.WithLabelValues("success").Observe(time.Since(startTime).Seconds())
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *MyAppReconciler) cleanupResources(ctx context.Context, myApp *myappv1.MyApp) error {
	log := ctrl.LoggerFrom(ctx)

	// Delete associated Deployment
	var deployment appv1.Deployment
	if err := r.Client.Get(ctx, types.NamespacedName{Name: myApp.Name + "-deployment", Namespace: myApp.Namespace}, &deployment); err == nil {
		log.Info("Deleting Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
		if err := r.Client.Delete(ctx, &deployment); err != nil {
			log.Error(err, "Failed to delete Deployment", "namespace", deployment.Namespace, "name", deployment.Name)
			return err
		}
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	// Delete associated PodDisruptionBudget
	var pdb policyv1.PodDisruptionBudget
	if err := r.Client.Get(ctx, types.NamespacedName{Name: myApp.Name + "-pdb", Namespace: myApp.Namespace}, &pdb); err == nil {
		log.Info("Deleting PodDisruptionBudget", "namespace", pdb.Namespace, "name", pdb.Name)
		if err := r.Client.Delete(ctx, &pdb); err != nil {
			log.Error(err, "Failed to delete PodDisruptionBudget", "namespace", pdb.Namespace, "name", pdb.Name)
			return err
		}
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func removeString(slice []string, str string) []string {
	newSlice := []string{}
	for _, s := range slice {
		if s != str {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log = log.Log.WithName("controller").WithName("MyApp")
	return ctrl.NewControllerManagedBy(mgr).
		For(&myappv1.MyApp{}).
		Owns(&appv1.Deployment{}).
		Complete(r)
}
