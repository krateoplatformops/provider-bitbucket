package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RepoParams struct {
	// Project: the project key.
	// +immutable
	Project string `json:"project"`

	// Name: the name of the repository.
	// +immutable
	Name string `json:"name"`

	// Private: whether the repository is private (default: true).
	// +optional
	// +immutable
	Private bool `json:"private,omitempty"`
}

type RepoObservation struct {
	// Project: the project key
	Project *string `json:"project,omitempty"`

	// RepoSlug: the repository name slug.
	RepoSlug *string `json:"repoSlug,omitempty"`
}

// A RepoSpec defines the desired state of a Repo.
type RepoSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RepoParams `json:"forProvider"`
}

// A RepoStatus represents the observed state of a Repo.
type RepoStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RepoObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Repo is a managed resource that represents a bitbucket repository
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".status.atProvider.project"
// +kubebuilder:printcolumn:name="SLUG",type="string",JSONPath=".status.atProvider.repoSlug"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",priority=1
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=1
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,krateo,bitbucket}
type Repo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepoSpec   `json:"spec"`
	Status RepoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepoList contains a list of Repo.
type RepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repo `json:"items"`
}
