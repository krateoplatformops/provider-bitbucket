package repopermissionuser

import (
	"context"
	"errors"
	"fmt"
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/krateoplatformops/provider-bitbucket/apis/repopermissionuser/v1alpha1"
	"github.com/krateoplatformops/provider-bitbucket/pkg/clients"
	"github.com/krateoplatformops/provider-bitbucket/pkg/clients/bitbucket"
	"github.com/krateoplatformops/provider-bitbucket/pkg/helpers"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	errNotRepoPermissionUser = "managed resource is not a repo permission user custom resource"

	reasonCannotCreate = "CannotCreateExternalResource"
	reasonCreated      = "CreatedExternalResource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RepoPermissionUserGroupKind)

	log := o.Logger.WithValues("controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RepoPermissionUserGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: mgr.GetEventRecorderFor(name),
			clientFn: bitbucket.NewClient,
		}),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.RepoPermissionUser{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
	clientFn func(opts *bitbucket.ClientOpts) *bitbucket.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.RepoPermissionUser)
	if !ok {
		return nil, errors.New(errNotRepoPermissionUser)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &external{
		kube: c.kube,
		log:  c.log,
		cli:  bitbucket.NewClient(cfg),
		rec:  c.recorder,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	kube client.Client
	log  logging.Logger
	cli  *bitbucket.Client
	rec  record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.RepoPermissionUser)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRepoPermissionUser)
	}

	spec := cr.Spec.ForProvider.DeepCopy()

	usr, err := e.cli.Repos().GetUserPermissions(bitbucket.UserPermissionOpts{
		ProjectKey: spec.Project,
		RepoSlug:   spec.RepoSlug,
		User:       spec.User,
	})
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if usr == nil {
		return managed.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	cr.Status.AtProvider.Project = helpers.StringPtr(spec.Project)
	cr.Status.AtProvider.RepoSlug = helpers.StringPtr(spec.RepoSlug)
	cr.Status.AtProvider.User = helpers.StringPtr(usr.User.Name)
	cr.Status.AtProvider.Permission = helpers.StringPtr(usr.Permission)

	isUpToDate := strings.HasSuffix(usr.Permission, strings.ToUpper(spec.Permission))

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: isUpToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.RepoPermissionUser)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRepoPermissionUser)
	}

	cr.SetConditions(xpv1.Creating())

	spec := cr.Spec.ForProvider.DeepCopy()

	repos := e.cli.Repos()
	err := repos.SetUserPermissions(bitbucket.UserPermissionOpts{
		ProjectKey: spec.Project,
		RepoSlug:   spec.RepoSlug,
		User:       spec.User,
		Permission: fmt.Sprintf("REPO_%s", strings.ToUpper(spec.Permission)),
	})
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	e.log.Debug("User permission created",
		"project", spec.Project,
		"user", spec.User,
		"slug", spec.RepoSlug,
		"perm", spec.Permission)
	e.rec.Event(cr, corev1.EventTypeNormal, reasonCreated, fmt.Sprintf("User permission '%s/%s/%s' created", spec.Project, spec.RepoSlug, spec.Permission))

	//meta.SetExternalName(cr, fmt.Sprintf("%s/%s", spec.Project, res.Slug))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.RepoPermissionUser)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRepoPermissionUser)
	}

	spec := cr.Spec.ForProvider.DeepCopy()

	repos := e.cli.Repos()
	err := repos.SetUserPermissions(bitbucket.UserPermissionOpts{
		ProjectKey: spec.Project,
		RepoSlug:   spec.RepoSlug,
		User:       spec.User,
		Permission: fmt.Sprintf("REPO_%s", strings.ToUpper(spec.Permission)),
	})
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	e.log.Debug("User permission updated",
		"project", spec.Project,
		"user", spec.User,
		"slug", spec.RepoSlug,
		"perm", spec.Permission)
	e.rec.Event(cr, corev1.EventTypeNormal, reasonCreated, fmt.Sprintf("User permission '%s/%s/%s' updated", spec.Project, spec.RepoSlug, spec.Permission))

	return managed.ExternalUpdate{}, nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.RepoPermissionUser)
	if !ok {
		return errors.New(errNotRepoPermissionUser)
	}

	cr.SetConditions(xpv1.Deleting())

	return e.cli.Repos().DeleteUserPermissions(bitbucket.UserPermissionOpts{
		ProjectKey: helpers.StringValue(cr.Status.AtProvider.Project),
		RepoSlug:   helpers.StringValue(cr.Status.AtProvider.RepoSlug),
		User:       helpers.StringValue(cr.Status.AtProvider.User),
	})
}
