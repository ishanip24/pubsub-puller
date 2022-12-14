/*
Copyright 2022 pc.

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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myoperatorv1 "example.com/m/api/v1"
)

// PubSubListenerReconciler reconciles a PubSubListener object
type PubSubListenerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.company.org,resources=pubsublisteners,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.company.org,resources=pubsublisteners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.company.org,resources=pubsublisteners/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *PubSubListenerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1: Load the PubSubListener by name
	var pubSubListener myoperatorv1.PubSubListener
	if err := r.Get(ctx, req.NamespacedName, &pubSubListener); err != nil {
		log.Error(err, "unable to fetch PubSubListener")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var mostRecentTime *time.Time // find the last run so we can update the status

	// We'll store the launch time in an annotation, so we'll reconstitute that from
	// the active jobs themselves.
	scheduledTimeForJob := time.Unix(*pubSubListener.Spec.PollingInterval, 0)
	if &scheduledTimeForJob != nil {
		if mostRecentTime == nil {
			mostRecentTime = &scheduledTimeForJob
		} else if mostRecentTime.Before(scheduledTimeForJob) {
			mostRecentTime = &scheduledTimeForJob
		}
	}
	if mostRecentTime != nil {
		pubSubListener.Status.LastScheduleTime = &metav1.Time{Time: *mostRecentTime}
	} else {
		pubSubListener.Status.LastScheduleTime = nil
	}

	if err := r.Status().Update(ctx, &pubSubListener); err != nil {
		log.Error(err, "unable to update PubSubListener status")
		return ctrl.Result{}, err
	}

	// 2: check if we're suspended
	if pubSubListener.Spec.Suspend != nil && *pubSubListener.Spec.Suspend {
		log.V(1).Info("Job suspended, skipping")
		return ctrl.Result{}, nil
	}

	// 3: get next scheduled run

	// Check if there is a deployment that matches the name of the pubsub subscription
	subscriptionPullers := pubSubListener.Status.SubscriptionPuller
	var found *appsv1.Deployment
	for _, subscriptionPuller := range subscriptionPullers {
		subName := subscriptionPuller.String()
		// convert the name conventions to match
		subName = strings.ToLower(subName)
		subName = strings.ReplaceAll(subName, "_", "")

		// Decide if one matches the pubsub name
		if subName == pubSubListener.Spec.SubscriptionName {
			if pubSubListener.Spec.Suspend == nil && !*pubSubListener.Spec.Suspend {
				found = &subscriptionPuller
			}
			break
		}
	}

	clientSet, clientSetErr := NewKubeClient()
	if clientSetErr != nil {
		return ctrl.Result{}, clientSetErr
	}
	// If the instance is active but is suspended, delete the deployment
	deploymentsClient := clientSet.AppsV1().Deployments(corev1.NamespaceDefault)
	if found != nil && pubSubListener.Spec.Suspend != nil && found.Status.Conditions[0].Type != appsv1.DeploymentAvailable {
		// set deletion timestamp on the suspended deployment
		deletePolicy := metav1.DeletePropagationForeground
		if err := deploymentsClient.Delete(ctx, found.Name, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); err != nil {
			panic(err)
		}
		log.V(1).Info("Deleted deployment.")
	}
	// Only create if deployment not found
	if found == nil && *pubSubListener.Spec.Suspend == false { // if subscription name not found, create deployment
		err := createSub(ctx, log, pubSubListener, "pull-test-results", "tmp-district-id")
		if err != nil {
			log.Error(err, "error creating subscription for ", pubSubListener.Spec.SubscriptionName)
		}
	}

	return ctrl.Result{}, nil
}

func createSub(ctx context.Context, log logr.Logger, pubSubListener myoperatorv1.PubSubListener, namespace string, districtID string) error {
	clientSet, clientSetErr := NewKubeClient()
	if clientSetErr != nil {
		return clientSetErr
	}
	nope := false
	yep := true
	defaultTerminationGracePeriodSeconds := int64(30)
	nobody := int64(65534)

	appsClient := clientSet.AppsV1().Deployments(namespace)
	cpuLimit := resource.MustParse("1")
	memoryLimit := resource.MustParse("1Gi")
	fileMode := int32(0o644)
	myDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pubsublistener",
			Namespace: namespace,
			Labels: map[string]string{
				"app/listener": "pubsublistener",
				"tier":         "backend",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app/listener": "pubsublistener",
					"tier":         "backend",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app/listener": "pubsublistener",
						"tier":         "backend",
					},
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: &yep,
					Containers: []corev1.Container{
						{
							Args: []string{"pull-topic", "-forever"},
							Env: []corev1.EnvVar{
								{
									Name:  "DEBUG",
									Value: "\"TRUE\"",
								},
								{
									Name:  "LOCAL_BASEPATH",
									Value: "/scratch",
								},
								{
									Name:  "KA_IS_DEV_SERVER",
									Value: "\"0\"",
								},
								{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/config/secret/service-account-credentials.json",
								},
							},
							Image:           "gcr.io/khan-internal-services/districts-jobs-roster:50e642a40dd5ab694b29029cde309c19c4609695",
							ImagePullPolicy: corev1.PullAlways,
							Name:            "pubsublistener",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged:               &nope,
								RunAsUser:                &nobody,
								RunAsGroup:               &nobody,
								ReadOnlyRootFilesystem:   &nope,
								AllowPrivilegeEscalation: &nope,
							},
							TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
							TerminationMessagePath:   "/dev/termination-log",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "service-account-credentials-vol",
									MountPath: "/config/secret",
								},
								{
									Name:      "scratch-vol",
									MountPath: "/scratch",
								},
							},
						},
					},
					DNSPolicy:                     corev1.DNSClusterFirst,
					EnableServiceLinks:            &yep,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					ServiceAccountName:            corev1.NamespaceDefault,
					TerminationGracePeriodSeconds: &defaultTerminationGracePeriodSeconds,
					Volumes: []corev1.Volume{
						{
							Name: "scratch-vol",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
						{
							Name: "service-account-credentials-vol",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "service-account-credentials",
									DefaultMode: &fileMode,
									Optional:    &nope,
								},
							},
						},
					},
				},
			},
		},
	}
	opts := metav1.CreateOptions{}
	result, err := appsClient.Create(ctx, myDeployment, opts)
	if err != nil {
		log.Error(err, "Unable to create job")
		return err
	}
	log.Info(fmt.Sprintf("Created job %q.\n", result.GetObjectMeta().GetName()))
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PubSubListenerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myoperatorv1.PubSubListener{}).
		Complete(r)
}
