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

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PubSubListenerSpec defines the desired state of PubSubListener
type PubSubListenerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Interval time in seconds to poll in Unix Timestamp format.
	// See https://www.unixtimestamp.com/
	PollingInterval *int64 `json:"pollinginterval"`

	// deployment should match the name of the pubsub subscription
	SubscriptionName string `json:"subscriptionname,omitempty"`

	//+kubebuilder:validation:Minimum=0

	// The number of successful finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	//+kubebuilder:validation:Minimum=0

	// The number of failed finished jobs to retain.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`
}

// PubSubListenerStatus defines the observed state of PubSubListener
type PubSubListenerStatus struct {
	// A list of pointers to current deployments
	// +optional
	SubscriptionPuller []appsv1.Deployment `json:"subscriptionpuller"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PubSubListener is the Schema for the pubsublisteners API
type PubSubListener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PubSubListenerSpec   `json:"spec,omitempty"`
	Status PubSubListenerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PubSubListenerList contains a list of PubSubListener
type PubSubListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PubSubListener `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PubSubListener{}, &PubSubListenerList{})
}
