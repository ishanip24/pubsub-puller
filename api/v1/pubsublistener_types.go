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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO Tmp implementation, not sure what goes into Deployment
type Deployment struct {
	SubscriptionName string `json:"subscriptionname,omitempty"`
}

// PubSubListenerSpec defines the desired state of PubSubListener
type PubSubListenerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// deployment should match the name of the pubsub subscription
	Subscription string `json:"subscription,omitempty"`
}

// PubSubListenerStatus defines the observed state of PubSubListener
type PubSubListenerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Deployments []Deployment `json:"deploymeny,omitempty"`
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
