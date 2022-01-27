/*

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgresSpec defines the desired state of Postgres
type PostgresSpec struct {
	DeploymentName string     `json:"deploymentName"`
	Image          string     `json:"image"`
	Replicas       *int32     `json:"replicas"`
	Users          []UserSpec `json:"users"`
	Databases      []string   `json:"databases"`
	Commands       []string   `json:"initcommands"`
}

// PostgresStatus defines the observed state of Postgres
type PostgresStatus struct {
	AvailableReplicas int32      `json:"availableReplicas"`
	ActionHistory     []string   `json:"actionHistory"`
	Users             []UserSpec `json:"users"`
	Databases         []string   `json:"databases"`
	VerifyCmd         string     `json:"verifyCommand"`
	ServiceIP         string     `json:"serviceIP"`
	ServicePort       string     `json:"servicePort"`
	Status            string     `json:"status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Postgres is the Schema for the postgres API
// +k8s:openapi-gen=true
type Postgres struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PostgresSpec   `json:"spec"`
	Status            PostgresStatus `json:"status,omitempty"`
}

type UserSpec struct {
	User     string `json:"username"`
	Password string `json:"password"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PostgresList contains a list of Postgres
type PostgresList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Postgres `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Postgres{}, &PostgresList{})
}
