/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dummyv1alpha1 "github.com/eupolemo/dummy-operator/api/v1alpha1"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=dummy.interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dummy.interview.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dummy.interview.com,resources=dummies/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	dummy := &dummyv1alpha1.Dummy{}
	err := r.Get(ctx, req.NamespacedName, dummy)

	logger.Info(dummy.Spec.Message)

	//dummyDeployment := &appsv1.Deployment{}

	if dummy.Spec.Message != dummy.Status.SpecEcho {
		dummy.Status.SpecEcho = dummy.Spec.Message
	}

	// TODO(user): your logic here

	found := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: dummy.Name, Namespace: dummy.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		pod, err := r.podForDummy(dummy)
		if err != nil {
			logger.Error(err, "Failed do define new Pod resource for Dummy")

			dummy.Status.PodStatus = "Pending"

			if err = r.Status().Update(ctx, dummy); err != nil {
				logger.Error(err, "Failed to update Dummy status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		logger.Info("Creating a new Pod",
			"Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		if err = r.Create(ctx, pod); err != nil {
			logger.Error(err, "Failed to create new Pod",
				"Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return ctrl.Result{}, err
		}

		dummy.Status.PodStatus = "Pending"

		if err = r.Status().Update(ctx, dummy); err != nil {
			logger.Error(err, "Failed to update Dummy status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Pod")
		return ctrl.Result{}, err
	}

	if found.Status.Phase == corev1.PodRunning {
		dummy.Status.PodStatus = "Running"
		logger.Info("Pod is Running!")
	} else {
		dummy.Status.PodStatus = "Pending"
		logger.Info("Pod is Pending!")
	}

	if err = r.Status().Update(ctx, dummy); err != nil {
		logger.Error(err, "Failed to update Dummy status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DummyReconciler) podForDummy(
	dummy *dummyv1alpha1.Dummy) (*corev1.Pod, error) {
	ls := labelsForDummy(dummy.Name)

	image, err := imageForDummy()
	if err != nil {
		return nil, err
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    ls,
			Name:      dummy.Name,
			Namespace: dummy.Namespace,
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &[]bool{true}[0],
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
			},
			Containers: []corev1.Container{{
				Image:           image,
				Name:            "dummy",
				ImagePullPolicy: corev1.PullIfNotPresent,
				SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot:             &[]bool{true}[0],
					RunAsUser:                &[]int64{1001}[0],
					AllowPrivilegeEscalation: &[]bool{false}[0],
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{
							"ALL",
						},
					},
				},
				Ports: []corev1.ContainerPort{{
					ContainerPort: 80,
					Name:          "dummy",
				}},
				Command: []string{"nginx", "-g", "daemon off;"},
			}},
		},
	}

	if err := ctrl.SetControllerReference(dummy, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}

func labelsForDummy(name string) map[string]string {
	var imageTag string
	image, err := imageForDummy()
	if err == nil {
		imageTag = strings.Split(image, ":")[1]
	}
	return map[string]string{"app.kubernetes.io/name": "Dummy",
		"app.kubernetes.io/instance":   name,
		"app.kubernetes.io/version":    imageTag,
		"app.kubernetes.io/part-of":    "dummy-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
}

func imageForDummy() (string, error) {
	var imageEnvVar = "DUMMY_IMAGE"
	image, found := os.LookupEnv(imageEnvVar)
	if !found {
		return "nginxinc/nginx-unprivileged:latest", nil
		//return "", fmt.Errorf("unable to find %s environment variable with the image", imageEnvVar)
	}
	return image, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dummyv1alpha1.Dummy{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
