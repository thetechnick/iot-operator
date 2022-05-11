package rollershutterrequests

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iotv1alpha1 "github.com/thetechnick/iot-operator/apis/iot/v1alpha1"
)

const requestHistoryLimit = 5

type RollerShutterRequestReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *RollerShutterRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iotv1alpha1.RollerShutterRequest{}).
		Complete(r)
}

func (r *RollerShutterRequestReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("rollershutterrequest", req.NamespacedName.String())
	defer log.Info("reconciled")

	rollerShutterRequestList := &iotv1alpha1.RollerShutterRequestList{}
	if err := r.List(ctx, rollerShutterRequestList); err != nil {
		return res, fmt.Errorf("listing RollerShutterRequests: %w", err)
	}

	requests := map[client.ObjectKey]int{}
	for _, req := range rollerShutterRequestList.Items {
		if !meta.IsStatusConditionTrue(req.Status.Conditions, iotv1alpha1.RollerShutterRequestCompleted) {
			// TODO: Add reference check to find and report invalid references
			continue
		}

		key := client.ObjectKey{
			Name:      req.Spec.RollerShutter.Name,
			Namespace: req.Namespace,
		}
		requests[key]++
		if requests[key] >= requestHistoryLimit {
			if err := r.Delete(ctx, &req); err != nil {
				return res, fmt.Errorf("garbage collecting RollerShutterRequest: %w", err)
			}
		}
	}

	return
}
