/*
Copyright 2023 The Crossplane Authors.

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

package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/feature"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	invv1alpha1 "github.com/rossigee/provider-btcpay/apis/invoice/v1alpha1"
	storev1alpha1 "github.com/rossigee/provider-btcpay/apis/store/v1alpha1"
	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/rossigee/provider-btcpay/apis"
	"github.com/rossigee/provider-btcpay/internal/controller"
	"github.com/rossigee/provider-btcpay/internal/tracing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
)

func main() {
	var (
		app                      = kingpin.New(filepath.Base(os.Args[0]), "BTCPay Server support for Crossplane.").DefaultEnvars()
		debug                    = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		leaderElection           = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		leaderElectionNS         = app.Flag("leader-election-namespace", "Namespace to use for leader election.").Default("crossplane-system").OverrideDefaultFromEnvar("LEADER_ELECTION_NAMESPACE").String()
		pollInterval             = app.Flag("poll", "How often individual resources will be checked for drift from the desired state").Short('p').Default("1m").Duration()
		maxReconcileRate         = app.Flag("max-reconcile-rate", "The global maximum rate per second at which resources may checked for drift from the desired state.").Default("10").Int()
		syncPeriod               = app.Flag("sync", "How often all resources will be double-checked for drift from the desired state.").Short('s').Default("1h").Duration()
		enableManagementPolicies = app.Flag("enable-management-policies", "Enable support for management policies.").Default("true").OverrideDefaultFromEnvar("ENABLE_MANAGEMENT_POLICIES").Bool()
		pollStateMetricInterval  = app.Flag("poll-state-metric", "State metric recording interval").Default("5s").Duration()
		metricsBindAddress       = app.Flag("metrics-bind-address", "The address the metrics endpoint binds to.").Default(":8080").String()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-btcpay"))

	shutdownTracing := tracing.Init("provider-btcpay")
	defer shutdownTracing(context.Background())
	// The controller-runtime runs with a no-op logger by default. Always set it
	// to ensure logs are displayed. In debug mode, it will be more verbose.
	ctrl.SetLogger(zl)

	log.Debug("Starting", "sync-period", syncPeriod.String())

	cfg, err := ctrl.GetConfig()
	if err != nil {
		kingpin.FatalIfError(err, "Cannot get API server rest config")
	}

	s := runtime.NewScheme()
	kingpin.FatalIfError(scheme.AddToScheme(s), "Cannot add k8s types to scheme")
	kingpin.FatalIfError(apis.AddToScheme(s), "Cannot add BTCPay APIs to scheme")

	mgr, err := ctrl.NewManager(ratelimiter.LimitRESTConfig(cfg, *maxReconcileRate), ctrl.Options{
		Scheme: s,
		Cache: cache.Options{
			SyncPeriod: syncPeriod,
		},
		LeaderElection:             *leaderElection,
		LeaderElectionID:           "crossplane-leader-election-provider-btcpay",
		LeaderElectionNamespace:    *leaderElectionNS,
		LeaderElectionResourceLock: "leases",
		LeaseDuration:              func() *time.Duration { d := 60 * time.Second; return &d }(),
		RenewDeadline:              func() *time.Duration { d := 50 * time.Second; return &d }(),
		Metrics: metricserver.Options{
			BindAddress: *metricsBindAddress,
		},
	})
	if err != nil {
		kingpin.FatalIfError(err, "Cannot create controller manager")
	}

	mrStateMetrics := statemetrics.NewMRStateMetrics()
	metrics.Registry.MustRegister(mrStateMetrics)

	mo := xpcontroller.MetricOptions{
		PollStateMetricInterval: *pollStateMetricInterval,
		MRStateMetrics:          mrStateMetrics,
	}

	o := xpcontroller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcileRate,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcileRate),
		Features:                &feature.Flags{},
		MetricOptions:           &mo,
	}

	if *enableManagementPolicies {
		o.Features.Enable(feature.EnableBetaManagementPolicies)
		log.Info("Beta feature enabled", "flag", feature.EnableBetaManagementPolicies)
	}

	if err := controller.Setup(mgr, o); err != nil {
		kingpin.FatalIfError(err, "Cannot setup BTCPay controllers")
	}

	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &storev1alpha1.StoreList{}, o.MetricOptions.PollStateMetricInterval)), "Cannot register state metrics for Store")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), o.Logger, o.MetricOptions.MRStateMetrics, &invv1alpha1.InvoiceList{}, o.MetricOptions.PollStateMetricInterval)), "Cannot register state metrics for Invoice")

	kingpin.FatalIfError(mgr.AddHealthzCheck("healthz", healthz.Ping), "Cannot add health check")
	kingpin.FatalIfError(mgr.AddReadyzCheck("readyz", healthz.Ping), "Cannot add ready check")

	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
