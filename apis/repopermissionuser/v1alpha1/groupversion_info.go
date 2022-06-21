// Package v1alpha1 contains the resources of the provider.
// +kubebuilder:object:generate=true
// +groupName=bitbucket.krateo.io
// +versionName=v1alpha1
package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "bitbucket.krateo.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

var (
	RepoPermissionUserKind             = reflect.TypeOf(RepoPermissionUser{}).Name()
	RepoPermissionUserGroupKind        = schema.GroupKind{Group: Group, Kind: RepoPermissionUserKind}.String()
	RepoPermissionUserKindAPIVersion   = RepoPermissionUserKind + "." + SchemeGroupVersion.String()
	RepoPermissionUserGroupVersionKind = SchemeGroupVersion.WithKind(RepoPermissionUserKind)
)

func init() {
	SchemeBuilder.Register(&RepoPermissionUser{}, &RepoPermissionUserList{})
}
