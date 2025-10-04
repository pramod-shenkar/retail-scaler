/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	crdv1 "retail-scalar/api/v1"
	"retail-scalar/utils"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	appsv1 "k8s.io/api/apps/v1"

	// autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScalarReconciler reconciles a Scalar object
type ScalarReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	org        *crdv1.Org
	queue      *utils.Client
	retryAfter time.Duration
}

// +kubebuilder:rbac:groups=vajra.com,resources=scalars,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vajra.com,resources=scalars/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vajra.com,resources=scalars/finalizers,verbs=update

func (r *ScalarReconciler) Reconcile(ctx context.Context, req ctr.Request) (ctr.Result, error) {

	if r.org == nil {
		var org crdv1.Org
		err := r.Get(ctx, req.NamespacedName, &org)
		if err != nil {
			slog.ErrorContext(ctx, "errro while getting org", "err", err)
			return ctr.Result{}, err
		}

		r.org = &org
	}

	if r.retryAfter == 0 {
		r.retryAfter = time.Second * r.org.Spec.LagCheckerDelay
	}

	if r.queue == nil {
		client, err := utils.NewClient(ctx, kafka.ConfigMap{
			// "bootstrap.servers": fmt.Sprintf(
			// 	"%v.%v.svc.cluster.local:%v",
			// 	r.org.Spec.Queue.Name,
			// 	r.org.Spec.Queue.Namespace,
			// 	r.org.Spec.Queue.Ports.Plaintext,
			// ),
			"bootstrap.servers": "localhost:9094",
			"group.id":          r.org.Spec.Consumers.GroupId,
		})
		if err != nil {
			slog.ErrorContext(ctx, "error while getting queue client instance", "err", err)
			return ctr.Result{RequeueAfter: r.retryAfter / 3}, err
		}

		r.queue = client
	}

	lag, err := r.queue.Lag(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting queue client instance", "err", err)
		return ctr.Result{RequeueAfter: r.retryAfter / 3}, err
	}

	var org crdv1.Org
	err = r.Get(ctx, req.NamespacedName, &org)
	if err != nil {
		slog.ErrorContext(ctx, "errro while getting org", "err", err)
		return ctr.Result{RequeueAfter: r.retryAfter / 3}, err
	}

	var topics = org.Spec.Consumers.Topics

	for _, topic := range topics {

		slog.Warn("found", "topic", topic, "lag", lag[topic.Name], "desired replicas", GetDesiredReplicas(topic, lag[topic.Name]))

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      topic.Name,
				Namespace: org.Spec.Consumers.Namespace,
			},
		}

		patch := []byte(fmt.Sprintf(`{"spec":{"replicas":%d}}`, GetDesiredReplicas(topic, lag[topic.Name])))

		err = r.Patch(ctx, deployment, client.RawPatch(types.MergePatchType, patch))
		if err != nil {
			slog.ErrorContext(ctx, "errro while updating consumer deployment", "err", err)
			return ctr.Result{RequeueAfter: r.retryAfter / 3}, err
		}
	}

	return ctr.Result{RequeueAfter: r.retryAfter}, nil
}

func GetDesiredReplicas(topic crdv1.Topic, lag int64) int32 {

	desired := int(math.Ceil(float64(lag) / float64(topic.MaxLag)))

	// at least on replica
	desired = max(desired, 1)

	// maintain min replica's needed
	desired = min(desired, topic.PartitionCount)

	return int32(desired)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalarReconciler) SetupWithManager(mgr ctr.Manager) error {
	return ctr.NewControllerManagedBy(mgr).
		For(&crdv1.Org{}).
		Named("scalar").
		Complete(r)
}
