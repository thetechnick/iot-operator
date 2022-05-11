---
title: API Reference
weight: 50
---

# IoT Operator API Reference

The IoT Operator APIs are an extension of the [Kubernetes API](https://kubernetes.io/docs/reference/using-api/api-overview/) using `CustomResourceDefinitions`.

## `iot.thetechnick.ninja`

The `iot.thetechnick.ninja` API group in contains all IoT related API objects.

* [RollerShutterRequest](#rollershutterrequestiotmanagedopenshiftiov1alpha1)
	* [RollerShutterRequestSpec](#rollershutterrequestspeciotmanagedopenshiftiov1alpha1)
	* [RollerShutterRequestStatus](#rollershutterrequeststatusiotmanagedopenshiftiov1alpha1)
* [RollerShutter](#rollershutteriotmanagedopenshiftiov1alpha1)
	* [RollerShutterEndpoint](#rollershutterendpointiotmanagedopenshiftiov1alpha1)
	* [RollerShutterSpec](#rollershutterspeciotmanagedopenshiftiov1alpha1)
	* [RollerShutterStatus](#rollershutterstatusiotmanagedopenshiftiov1alpha1)

### RollerShutterRequest.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta) | false |
| spec |  | [RollerShutterRequestSpec.iot.managed.openshift.io/v1alpha1](#rollershutterrequestspeciotmanagedopenshiftiov1alpha1) | false |
| status |  | [RollerShutterRequestStatus.iot.managed.openshift.io/v1alpha1](#rollershutterrequeststatusiotmanagedopenshiftiov1alpha1) | false |

[Back to Group]()

### RollerShutterRequestSpec.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| position | Desired position for the shutter. | int.iot.managed.openshift.io/v1alpha1 | true |
| rollerShutter |  | corev1.LocalObjectReference | true |

[Back to Group]()

### RollerShutterRequestStatus.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |
| conditions | Conditions is a list of status conditions ths object is in. | []metav1.Condition | false |
| phase |  | RollerShutterRequestPhase.iot.managed.openshift.io/v1alpha1 | false |

[Back to Group]()

### RollerShutter.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#objectmeta-v1-meta) | false |
| spec |  | [RollerShutterSpec.iot.managed.openshift.io/v1alpha1](#rollershutterspeciotmanagedopenshiftiov1alpha1) | false |
| status |  | [RollerShutterStatus.iot.managed.openshift.io/v1alpha1](#rollershutterstatusiotmanagedopenshiftiov1alpha1) | false |

[Back to Group]()

### RollerShutterEndpoint.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| url | URL to contact the device under. | string | true |

[Back to Group]()

### RollerShutterSpec.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| deviceType | Endpoint device type. | string | true |
| endpoint |  | [RollerShutterEndpoint.iot.managed.openshift.io/v1alpha1](#rollershutterendpointiotmanagedopenshiftiov1alpha1) | true |

[Back to Group]()

### RollerShutterStatus.iot.managed.openshift.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |
| conditions | Conditions is a list of status conditions ths object is in. | []metav1.Condition | false |
| phase |  | RollerShutterPhase.iot.managed.openshift.io/v1alpha1 | false |
| position | Recorded position in percentage open. 100 = completely open, 0 = completely closed. | int.iot.managed.openshift.io/v1alpha1 | true |
| power | Power consumption in Watts. | int.iot.managed.openshift.io/v1alpha1 | true |

[Back to Group]()
