package rollershutters

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	// Determine Client
	dt := rollerShutter.Spec.DeviceType
	switch dt {
	case shelly25Roller:
		if err := r.reconcileShelly25Roller(ctx, rollerShutter); err != nil {
			return res, fmt.Errorf("reconciling %s: %w", dt, err)
		}
	default:
		meta.SetStatusCondition(&rollerShutter.Status.Conditions, metav1.Condition{
			Type:    iotv1alpha1.RollerShutterReachable,
			Status:  metav1.ConditionFalse,
			Reason:  "UnkownDeviceType",
			Message: fmt.Sprintf("Unkown device type %q, must be one of: [%s]", dt, shelly25Roller),
		})

		meta.RemoveStatusCondition(&rollerShutter.Status.Conditions, iotv1alpha1.RollerShutterAtPosition)
	}

	rollerShutter.Status.ObservedGeneration = rollerShutter.Generation
	if err := r.Status().Update(ctx, rollerShutter); err != nil {
		return res, fmt.Errorf("updating RollerShutter status: %w", err)
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

func (r *RollerShutterReconciler) reconcileShelly25Roller(ctx context.Context, rollerShutter *iotv1alpha1.RollerShutter) error {
	c := shelly25rollerclient.NewClient(clients.WithEndpoint(rollerShutter.Spec.Endpoint.URL))

	status, err := c.ToPosition(ctx, rollerShutter.Spec.Position)
	if err != nil {
		return fmt.Errorf("commanding to position: %w", err)
	}

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

	if status.CurrentPos == rollerShutter.Spec.Position &&
		status.State == shelly25rollerclient.StateStop {
		meta.SetStatusCondition(&rollerShutter.Status.Conditions, metav1.Condition{
			Type:    iotv1alpha1.RollerShutterAtPosition,
			Status:  metav1.ConditionTrue,
			Reason:  "AtPosition",
			Message: "Position reached",
		})
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

	meta.SetStatusCondition(&rollerShutter.Status.Conditions, metav1.Condition{
		Type:    iotv1alpha1.RollerShutterAtPosition,
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})

	return nil
}
