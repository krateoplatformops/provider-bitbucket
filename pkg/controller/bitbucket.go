package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"

	"github.com/krateoplatformops/provider-bitbucket/pkg/controller/config"
	"github.com/krateoplatformops/provider-bitbucket/pkg/controller/repo"
	"github.com/krateoplatformops/provider-bitbucket/pkg/controller/repopermissionuser"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		config.Setup,
		repo.Setup,
		repopermissionuser.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
