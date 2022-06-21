package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RepoPermissionUserParams struct {
	// Project: the project key.
	// +immutable
	Project string `json:"project"`

	// RepoSlug: slug format of repository name.
	// +immutable
	RepoSlug string `json:"repoSlug"`

	// User: the user to grant permission.
	// +immutable
	User string `json:"user"`

	// Permission: the permission granted to the user (REPO_READ, REPO_WRITE, REPO_ADMIN).
	// +immutable
	Permission string `json:"permission"`
}

type RepoPersmissionUserObservation struct {
	// Project: the project key
	Project *string `json:"project,omitempty"`

	// RepoSlug: the repository name slug.
	RepoSlug *string `json:"repoSlug,omitempty"`

	// User: the user to grant permission.
	User *string `json:"user,omitempty"`

	// Permission: the permission granted to the user.
	Permission *string `json:"permission,omitempty"`
}

// A RepoUserPermissionSpec defines the desired state of a RepoPermissionUser.
type RepoPermissionUserSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RepoPermissionUserParams `json:"forProvider"`
}

// A RepoPermissionUserStatus represents the observed state of a Repo.
type RepoPermissionUserStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RepoPersmissionUserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Repo is a managed resource that represents a bitbucket repository
// +kubebuilder:printcolumn:name="PROJECT",type="string",JSONPath=".status.atProvider.project"
// +kubebuilder:printcolumn:name="SLUG",type="string",JSONPath=".status.atProvider.repoSlug"
// +kubebuilder:printcolumn:name="USER",type="string",JSONPath=".status.atProvider.user"
// +kubebuilder:printcolumn:name="PERM",type="string",JSONPath=".status.atProvider.permission"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",priority=1
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status",priority=1
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,krateo,bitbucket}
type RepoPermissionUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepoPermissionUserSpec   `json:"spec"`
	Status RepoPermissionUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepoList contains a list of Repo.
type RepoPermissionUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RepoPermissionUser `json:"items"`
}
