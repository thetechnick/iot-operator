package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	iotapis "github.com/thetechnick/iot-operator/apis"
	"github.com/thetechnick/iot-operator/internal/controllers/rollershutterrequests"
	"github.com/thetechnick/iot-operator/internal/controllers/rollershutters"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = iotapis.AddToScheme(scheme)
}

type options struct {
	metricsAddr           string
	pprofAddr             string
	enableLeaderElection  bool
	enableMetricsRecorder bool
	probeAddr             string
}

func parseFlags() *options {
	opts := &options{}

	flag.StringVar(&opts.metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&opts.pprofAddr, "pprof-addr", "", "The address the pprof web endpoint binds to.")
	flag.BoolVar(&opts.enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&opts.probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.BoolVar(&opts.enableMetricsRecorder, "enable-metrics-recorder", true, "Enable recording Addon Metrics")
	flag.Parse()

	return opts
}

func initReconcilers(mgr ctrl.Manager) error {
	rollerShutterReconciler := &rollershutters.RollerShutterReconciler{
		Client:                 mgr.GetClient(),
		Log:                    ctrl.Log.WithName("controllers").WithName("RollerShutter"),
		Scheme:                 mgr.GetScheme(),
		DefaultRequeueInterval: time.Second * 30,
		MovingRequeueInterval:  time.Second * 2,
	}

	if err := rollerShutterReconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create RollerShutter controller: %w", err)
	}

	rollerShutterRequestReconciler := &rollershutterrequests.RollerShutterRequestReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("RollerShutterRequest"),
		Scheme: mgr.GetScheme(),
	}

	if err := rollerShutterRequestReconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create RollerShutterRequest controller: %w", err)
	}
	return nil
}

func initPprof(mgr ctrl.Manager, addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s := &http.Server{Addr: addr, Handler: mux}
	err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		errCh := make(chan error)
		defer func() {
			for range errCh {
			} // drain errCh for GC
		}()
		go func() {
			defer close(errCh)
			errCh <- s.ListenAndServe()
		}()

		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
			s.Close()
			return nil
		}
	}))
	if err != nil {
		setupLog.Error(err, "unable to create pprof server")
		os.Exit(1)
	}
}

func setup() error {
	opts := parseFlags()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         opts.metricsAddr,
		HealthProbeBindAddress:     opts.probeAddr,
		Port:                       9443,
		LeaderElectionResourceLock: "leases",
		LeaderElection:             opts.enableLeaderElection,
		LeaderElectionID:           "8a4hp84a6s.iot-operator-lock",
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	// PPROF
	if len(opts.pprofAddr) > 0 {
		initPprof(mgr, opts.pprofAddr)
	}

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	if err := initReconcilers(mgr); err != nil {
		return fmt.Errorf("init reconcilers: %w", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}
	return nil
}

func main() {
	if err := setup(); err != nil {
		setupLog.Error(err, "setting up manager")
		os.Exit(1)
	}
}
