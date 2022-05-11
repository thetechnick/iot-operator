package rollershutters

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	iotv1alpha1 "github.com/thetechnick/iot-operator/apis/iot/v1alpha1"
	"github.com/thetechnick/iot-operator/internal/clients"
	"github.com/thetechnick/iot-operator/internal/clients/shelly25rollerclient"
)

type RollerShutterReconciler struct {
	client.Client
	Log                    logr.Logger
	Scheme                 *runtime.Scheme
	DefaultRequeueInterval time.Duration
	MovingRequeueInterval  time.Duration
}

func (r *RollerShutterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iotv1alpha1.RollerShutter{}).
		Watches(
			&source.Kind{
				Type: &iotv1alpha1.RollerShutterRequest{},
			},
			handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
				req := o.(*iotv1alpha1.RollerShutterRequest)
				return []reconcile.Request{
					{
						NamespacedName: client.ObjectKey{
							Name:      req.Spec.RollerShutter.Name,
							Namespace: req.Namespace,
						},
					},
				}
			}),
		).
		Complete(r)
}

const shelly25Roller = "Shelly25Roller"

func (r *RollerShutterReconciler) Reconcile(
	ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("rollershutter", req.NamespacedName.String())
	defer log.Info("reconciled")

	rollerShutter := &iotv1alpha1.RollerShutter{}
	if err := r.Get(ctx, req.NamespacedName, rollerShutter); err != nil {
		return res, client.IgnoreNotFound(err)
	}

	// List requests
	rollerShutterRequestList := &iotv1alpha1.RollerShutterRequestList{}
	if err := r.List(ctx, rollerShutterRequestList, client.InNamespace(rollerShutter.Namespace)); err != nil {
		return res, fmt.Errorf("listing RollerShutterRequests in namespace %s: %w", rollerShutter.Namespace, err)
	}
	var filteredRollerShutterRequests sortRequestByCreationTimestamp
	for _, req := range rollerShutterRequestList.Items {
		if req.Spec.RollerShutter.Name != rollerShutter.Name {
			continue
		}

		if meta.IsStatusConditionTrue(req.Status.Conditions, iotv1alpha1.RollerShutterRequestCompleted) {
			continue
		}

		filteredRollerShutterRequests = append(filteredRollerShutterRequests, req)
	}
	sort.Sort(filteredRollerShutterRequests)
	var request *iotv1alpha1.RollerShutterRequest
	if len(filteredRollerShutterRequests) > 0 {
		request = &filteredRollerShutterRequests[0]
	}

	// Determine Client
	dt := rollerShutter.Spec.DeviceType
	switch dt {
	case shelly25Roller:
		if err := r.reconcileShelly25Roller(ctx, rollerShutter, request); err != nil {
			return res, fmt.Errorf("reconciling %s: %w", dt, err)
		}
	default:
		meta.SetStatusCondition(&rollerShutter.Status.Conditions, metav1.Condition{
			Type:    iotv1alpha1.RollerShutterReachable,
			Status:  metav1.ConditionFalse,
			Reason:  "UnkownDeviceType",
			Message: fmt.Sprintf("Unkown device type %q, must be one of: [%s]", dt, shelly25Roller),
		})
	}

	// Handle status
	rollerShutter.Status.ObservedGeneration = rollerShutter.Generation
	if err := r.Status().Update(ctx, rollerShutter); err != nil {
		return res, fmt.Errorf("updating RollerShutter status: %w", err)
	}

	if request != nil {
		request.Status.ObservedGeneration = request.Generation
		if err := r.Status().Update(ctx, request); err != nil {
			return res, fmt.Errorf("updating RollerShutterRequest status: %w", err)
		}
	}

	if rollerShutter.Status.Phase == iotv1alpha1.RollerShutterPhaseIdle {
		// always get a new status every now and then
		res.RequeueAfter = r.DefaultRequeueInterval
	} else {
		// poll a bit more frequently while moving
		res.RequeueAfter = r.MovingRequeueInterval
	}
	return
}

func (r *RollerShutterReconciler) reconcileShelly25Roller(
	ctx context.Context, rollerShutter *iotv1alpha1.RollerShutter,
	req *iotv1alpha1.RollerShutterRequest,
) error {
	c := shelly25rollerclient.NewClient(clients.WithEndpoint(rollerShutter.Spec.Endpoint.URL))

	status, err := c.Status(ctx)
	if err != nil {
		return fmt.Errorf("reading status: %w", err)
	}

	// Status handling
	meta.SetStatusCondition(&rollerShutter.Status.Conditions, metav1.Condition{
		Type:    iotv1alpha1.RollerShutterReachable,
		Status:  metav1.ConditionTrue,
		Reason:  "Connected",
		Message: "connected to device",
	})

	rollerShutter.Status.Position = status.CurrentPos
	rollerShutter.Status.Power = int(status.Power)

	switch status.State {
	case shelly25rollerclient.StateClose:
		rollerShutter.Status.Phase = iotv1alpha1.RollerShutterPhaseClosing
	case shelly25rollerclient.StateOpen:
		rollerShutter.Status.Phase = iotv1alpha1.RollerShutterPhaseOpening
	// case shelly25rollerclient.StateStop:
	default:
		rollerShutter.Status.Phase = iotv1alpha1.RollerShutterPhaseIdle
	}

	// Request handling
	if req != nil {
		if req.Spec.Position != status.CurrentPos {
			status, err = c.ToPosition(ctx, req.Spec.Position)
			if err != nil {
				return fmt.Errorf("commanding to position: %w", err)
			}
		}

		if status.State == shelly25rollerclient.StateStop {
			// Move finished
			meta.SetStatusCondition(&req.Status.Conditions, metav1.Condition{
				Type:    iotv1alpha1.RollerShutterRequestCompleted,
				Status:  metav1.ConditionTrue,
				Reason:  "AtPosition",
				Message: "position reached",
			})
			req.Status.Phase = iotv1alpha1.RollerShutterRequestPhaseCompleted
			return nil
		}

		reason := "Moving"
		message := "moving shutter to position"
		if status.State == shelly25rollerclient.StateStop {
			switch status.StopReason {
			case shelly25rollerclient.StopReasonObstacle:
				reason = "Obstacle"
				message = "obstacle detected, stopped movement"
			case shelly25rollerclient.StopReasonSafetySwitch:
				reason = "SafetySwitch"
				message = "safety switch triggered"
			case shelly25rollerclient.StopReasonOverpower:
				reason = "Overpower"
				message = "overpower detected, stopped movement"
			}
		}

		meta.SetStatusCondition(&req.Status.Conditions, metav1.Condition{
			Type:    iotv1alpha1.RollerShutterRequestCompleted,
			Status:  metav1.ConditionFalse,
			Reason:  reason,
			Message: message,
		})
		req.Status.Phase = iotv1alpha1.RollerShutterRequestPhaseMoving
	}

	return nil
}

type sortRequestByCreationTimestamp []iotv1alpha1.RollerShutterRequest

func (p sortRequestByCreationTimestamp) Len() int {
	return len(p)
}

func (p sortRequestByCreationTimestamp) Less(i, j int) bool {
	return p[i].GetCreationTimestamp().UTC().Before(p[j].GetCreationTimestamp().UTC())
}

func (p sortRequestByCreationTimestamp) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
