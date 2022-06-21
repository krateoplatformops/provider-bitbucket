package repo

import (
	"context"
	"errors"
	"fmt"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	v1alpha1 "github.com/krateoplatformops/provider-bitbucket/apis/repo/v1alpha1"
	"github.com/krateoplatformops/provider-bitbucket/pkg/clients"
	"github.com/krateoplatformops/provider-bitbucket/pkg/clients/bitbucket"

	"github.com/Machiel/slugify"
)

const (
	errNotRepo = "managed resource is not a repo custom resource"

	reasonCannotCreate = "CannotCreateExternalResource"
	reasonCreated      = "CreatedExternalResource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RepoGroupKind)

	log := o.Logger.WithValues("controller", name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RepoGroupVersionKind),
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
		For(&v1alpha1.Repo{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
	clientFn func(opts *bitbucket.ClientOpts) *bitbucket.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Repo)
	if !ok {
		return nil, errors.New(errNotRepo)
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
	cr, ok := mg.(*v1alpha1.Repo)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRepo)
	}

	spec := cr.Spec.ForProvider.DeepCopy()
	slug := slugify.Slugify(spec.Name)
	ok, err := e.cli.Repos().Exists(spec.Project, slug)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if ok {
		e.log.Debug("Repo already exists", "project", spec.Project, "name", spec.Name)
		e.rec.Event(cr, corev1.EventTypeNormal, "AlredyExists", fmt.Sprintf("Repo '%s/%s' already exists", spec.Project, spec.Name))

		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	e.log.Debug("Repo does not exists", "org", spec.Project, "name", spec.Name)

	return managed.ExternalObservation{
		ResourceExists:   false,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Repo)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRepo)
	}

	cr.SetConditions(xpv1.Creating())

	spec := cr.Spec.ForProvider.DeepCopy()

	repos := e.cli.Repos()
	res, err := repos.Create(bitbucket.CreateRepoOpts{
		Name:       spec.Name,
		Public:     !spec.Private,
		ProjectKey: spec.Project,
	})
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	e.log.Debug("Repo created", "project", spec.Project, "name", spec.Name, "slug", res.Slug)
	e.rec.Event(cr, corev1.EventTypeNormal, reasonCreated, fmt.Sprintf("Repo '%s/%s' created", spec.Project, spec.Name))

	err = repos.Init(bitbucket.RepoInitOpts{
		ProjectKey: spec.Project,
		RepoSlug:   res.Slug,
		Title:      res.Description,
	})
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	e.log.Debug("Repo initialized", "project", spec.Project, "name", spec.Name, "slug", res.Slug)
	e.rec.Event(cr, corev1.EventTypeNormal, reasonCreated, fmt.Sprintf("Repo '%s/%s' initialized", spec.Project, spec.Name))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}
