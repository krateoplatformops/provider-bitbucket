package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/krateoplatformops/provider-bitbucket/apis/v1alpha1"
	"github.com/krateoplatformops/provider-bitbucket/pkg/clients/bitbucket"
	"github.com/krateoplatformops/provider-bitbucket/pkg/helpers"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultMaxIdleConnections = 5
	defaultResponseTimeout    = 30 * time.Second
	defaultConnectionTimeout  = 15 * time.Second
)

// GetConfig constructs a CreateOpts configuration that
// can be used to authenticate to the git API provider by the ReST client
func GetConfig(ctx context.Context, c client.Client, mg resource.Managed) (*bitbucket.ClientOpts, error) {
	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("providerConfigRef is not given")
	}
}

// UseProviderConfig to produce a config that can be used to create an ArgoCD client.
func UseProviderConfig(ctx context.Context, k client.Client, mg resource.Managed) (*bitbucket.ClientOpts, error) {
	pc := &v1alpha1.ProviderConfig{}
	err := k.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewProviderConfigUsageTracker(k, &v1alpha1.ProviderConfigUsage{})
	err = t.Track(ctx, mg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	if s := pc.Spec.Credentials.Source; s != xpv1.CredentialsSourceSecret {
		return nil, fmt.Errorf("credentials source %s is not currently supported", s)
	}

	csr := pc.Spec.Credentials.SecretRef
	if csr == nil {
		return nil, fmt.Errorf("no credentials secret referenced")
	}

	token, err := helpers.GetSecret(ctx, k, csr.DeepCopy())
	if err != nil {
		return nil, err
	}

	opts := &bitbucket.ClientOpts{
		ApiBaseUrl: pc.Spec.ApiUrl,
		Token:      token,
	}

	transport := http.DefaultTransport

	insecure := helpers.BoolValue(helpers.BoolOrDefault(pc.Spec.Insecure, false))
	if insecure {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		transport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSClientConfig:       tlsConfig,
			MaxIdleConnsPerHost:   defaultMaxIdleConnections,
			ResponseHeaderTimeout: defaultResponseTimeout,
		}
	}

	verbose := helpers.IsBoolPtrEqualToBool(pc.Spec.Verbose, true)
	if verbose {
		transport = &verboseTracer{transport}
	}

	opts.HttpClient = &http.Client{
		Transport: transport,
		Timeout:   defaultConnectionTimeout + defaultResponseTimeout,
	}

	return opts, nil
}

// verboseTracer implements http.RoundTripper.  It prints each request and
// response/error to os.Stderr.  WARNING: this may output sensitive information
// including bearer tokens.
type verboseTracer struct {
	http.RoundTripper
}

// RoundTrip calls the nested RoundTripper while printing each request and
// response/error to os.Stderr on either side of the nested call.  WARNING: this
// may output sensitive information including bearer tokens.
func (t *verboseTracer) RoundTrip(req *http.Request) (*http.Response, error) {
	// Dump the request to os.Stderr.
	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	os.Stderr.Write(b)
	os.Stderr.Write([]byte{'\n'})

	// Call the nested RoundTripper.
	resp, err := t.RoundTripper.RoundTrip(req)

	// If an error was returned, dump it to os.Stderr.
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return resp, err
	}

	// Dump the response to os.Stderr.
	b, err = httputil.DumpResponse(resp, req.URL.Query().Get("watch") != "true")
	if err != nil {
		return nil, err
	}
	os.Stderr.Write(b)
	os.Stderr.Write([]byte{'\n'})

	return resp, err
}
