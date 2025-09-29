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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	ctr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	cli "sigs.k8s.io/controller-runtime/pkg/client"

	crdv1 "retail-scalar/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OrgReconciler reconciles a Org object
type OrgReconciler struct {
	cli.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.vajra.com,resources=orgs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.vajra.com,resources=orgs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.vajra.com,resources=orgs/finalizers,verbs=update

/*
func (r *OrgReconciler) Reconcile(ctx context.Context, req ctr.Request) (ctr.Result, error) {

	var org crdv1.Org
	err := r.Get(ctx, req.NamespacedName, &org)
	if err != nil {
		slog.ErrorContext(ctx, "errro while getting org", "err", err)
		return ctr.Result{}, err
	}

	var queueNs = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: org.Spec.Queue.Namespace,
		},
	}
	err = r.CreateIfNotExist(ctx, types.NamespacedName{Name: org.Spec.Queue.Namespace}, &queueNs)
	if err != nil {
		slog.ErrorContext(ctx, "error while CreateIfNotExist queueNs", "err", err)
		return ctr.Result{}, err
	}

	var queueSerice = QueueService(org)
	err = r.CreateIfNotExist(ctx, types.NamespacedName{Name: org.Spec.Queue.Name, Namespace: org.Spec.Queue.Namespace}, queueSerice)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting service", "err", err)
		return ctr.Result{}, err
	}

	var queueDeployment = QueueDeployment(org)
	var queueKey = types.NamespacedName{Name: org.Spec.Queue.Name, Namespace: org.Spec.Queue.Namespace}

	err = r.CreateIfNotExist(ctx, queueKey, queueDeployment)
	if err != nil {
		slog.ErrorContext(ctx, "error while CreateIfNotExist queueNs", "err", err)
		return ctr.Result{}, err
	}

	err = wait.ExponentialBackoff(wait.Backoff{Duration: time.Second, Steps: 10}, func() (done bool, err error) {
		err = r.Get(ctx, queueKey, queueDeployment)
		if err != nil {
			slog.WarnContext(ctx, "error while", "err", err)
			return false, err
		}

		if queueDeployment.Status.AvailableReplicas >= *org.Spec.Queue.Replicas {
			slog.WarnContext(ctx, "tried", "got", queueDeployment.Status.AvailableReplicas >= *org.Spec.Queue.Replicas)
		}
		return queueDeployment.Status.AvailableReplicas >= *org.Spec.Queue.Replicas, nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "error while checking queue replicas", "err", err)
		return ctr.Result{}, err
	}

	// err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 4*time.Second, false, func(ctx context.Context) (done bool, err error) {

	// 	var queuePods corev1.PodList
	// 	err = r.List(ctx, &queuePods,
	// 		client.InNamespace(queueDeployment.Namespace),
	// 		client.MatchingLabels(queueDeployment.Spec.Selector.MatchLabels),
	// 	)
	// 	if err != nil {
	// 		slog.WarnContext(ctx, "error while", "err", err)
	// 		return false, err
	// 	}

	// 	var failed int
	// 	for _, pod := range queuePods.Items {
	// 		if pod.Status.Phase != corev1.PodRunning {
	// 			failed++
	// 		}
	// 	}

	// 	if failed > 0 {
	// 		return false, nil
	// 	}

	// 	return true, nil
	// })

	// if err != nil {
	// 	slog.ErrorContext(ctx, "error while checking pods for queue deployment", "err", err)
	// 	return ctr.Result{}, nil
	// }

	var topics = org.Spec.Topics

	var consumerNs = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "consumer",
		},
	}
	err = r.CreateIfNotExist(ctx, types.NamespacedName{Name: "consumer"}, &consumerNs)
	if err != nil {
		slog.ErrorContext(ctx, "error while CreateIfNotExist consumerNs", "err", err)
		return ctr.Result{}, err
	}

	for _, topic := range topics {

		var consumerKey = types.NamespacedName{Name: topic, Namespace: "consumer"}
		err = r.CreateIfNotExist(ctx, consumerKey, ConsumerDeployment(org, topic))
		if err != nil {
			slog.ErrorContext(ctx, "error while CreateIfNotExist consumer-"+topic, "err", err)
			return ctr.Result{}, err
		}
	}

	var errs = make(chan error, len(topics))
	var podCreateErr error
	go func() {
		for err := range errs {
			slog.WarnContext(ctx, "error while", "err", err)
			podCreateErr = errrs.Join(podCreateErr, err)
		}
	}()

	var wg = sync.WaitGroup{}
	wg.Add(len(topics))
	workqueue.ParallelizeUntil(ctx, min(10, len(topics)), len(topics), func(piece int) {

		defer wg.Done()

		var consumerKey = types.NamespacedName{Name: org.Spec.Queue.Name, Namespace: org.Spec.Queue.Namespace}
		err = r.CreateIfNotExist(ctx, consumerKey, ConsumerDeployment(org, topics[piece]))
		if err != nil {
			slog.ErrorContext(ctx, "error while CreateIfNotExist consumer-"+topics[piece], "err", err)
			errs <- err
		}
	})

	wg.Wait()
	close(errs)

	if podCreateErr != nil {
		slog.ErrorContext(ctx, "error while create consumer deployments", "err", err)
		return ctr.Result{}, podCreateErr
	}

	errs = make(chan error, len(topics))
	var deploymentGettingErr error
	go func() {
		for err := range errs {
			deploymentGettingErr = errrs.Join(deploymentGettingErr, err)
		}
	}()

	wg = sync.WaitGroup{}
	wg.Add(len(topics))
	workqueue.ParallelizeUntil(ctx, min(3, len(topics)), len(topics), func(piece int) {

		defer wg.Done()

		var deployment appsv1.Deployment
		err = wait.ExponentialBackoff(wait.Backoff{Duration: time.Second, Steps: 3}, func() (done bool, err error) {
			err = r.Get(ctx, types.NamespacedName{Name: topics[piece], Namespace: "consumer"}, &deployment)
			if err != nil {
				return false, err
			}
			return deployment.Status.ReadyReplicas > 0, nil
		})
		if err != nil {
			slog.ErrorContext(ctx, "error while CreateIfNotExist consumer-"+topics[piece], "err", err)
			errs <- err
		}
	})

	wg.Wait()
	close(errs)

	if deploymentGettingErr != nil {
		slog.ErrorContext(ctx, "error while getting consumer deployments", "err", err)
		return ctr.Result{}, deploymentGettingErr
	}

	return ctr.Result{}, nil
}
*/

func (r *OrgReconciler) Reconcile(ctx context.Context, req ctr.Request) (ctr.Result, error) {

	var org crdv1.Org
	err := r.Get(ctx, req.NamespacedName, &org)
	if err != nil {
		slog.ErrorContext(ctx, "errro while getting org", "err", err)
		return ctr.Result{}, err
	}

	var queueNs = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: org.Spec.Queue.Namespace,
		},
	}
	err = r.CreateIfNotExist(ctx, types.NamespacedName{Name: org.Spec.Queue.Namespace}, &queueNs)
	if err != nil {
		slog.ErrorContext(ctx, "error while CreateIfNotExist queueNs", "err", err)
		return ctr.Result{}, err
	}

	var queueSerice = QueueService(org)
	err = r.CreateIfNotExist(ctx, types.NamespacedName{Name: org.Spec.Queue.Name, Namespace: org.Spec.Queue.Namespace}, queueSerice)
	if err != nil {
		slog.ErrorContext(ctx, "error while getting service", "err", err)
		return ctr.Result{}, err
	}

	var queueDeployment = QueueDeployment(org)
	var queueKey = types.NamespacedName{Name: org.Spec.Queue.Name, Namespace: org.Spec.Queue.Namespace}

	err = r.CreateIfNotExist(ctx, queueKey, queueDeployment)
	if err != nil {
		slog.ErrorContext(ctx, "error while CreateIfNotExist queueNs", "err", err)
		return ctr.Result{}, err
	}

	var isSuccess bool
	for range 3 {
		time.Sleep(time.Second * 2)

		err = r.Get(ctx, queueKey, queueDeployment)
		if err != nil {
			slog.WarnContext(ctx, "error while", "err", err)
			continue
		}

		if queueDeployment.Status.AvailableReplicas == *org.Spec.Queue.Replicas {
			isSuccess = true
			break
		}

		slog.WarnContext(ctx, "retrying as found", "AvailableReplicas", queueDeployment.Status.AvailableReplicas, "expected", *org.Spec.Queue.Replicas)
	}

	if !isSuccess {
		slog.ErrorContext(ctx, "retries exaused")
		return ctr.Result{}, err
	}

	var topics = org.Spec.Consumers.Topics

	var consumerNs = corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: org.Spec.Consumers.Namespace,
		},
	}
	err = r.CreateIfNotExist(ctx, types.NamespacedName{Name: org.Spec.Consumers.Namespace}, &consumerNs)
	if err != nil {
		slog.ErrorContext(ctx, "error while CreateIfNotExist consumerNs", "err", err)
		return ctr.Result{}, err
	}

	for _, topic := range topics {

		var consumerKey = types.NamespacedName{Name: topic, Namespace: org.Spec.Consumers.Namespace}
		err = r.CreateIfNotExist(ctx, consumerKey, ConsumerDeployment(org, topic))
		if err != nil {
			slog.ErrorContext(ctx, "error while CreateIfNotExist consumer-"+topic, "err", err)
			return ctr.Result{}, err
		}
	}

	for piece := range topics {

		var isSuccess bool
		for range 3 {
			time.Sleep(time.Second * 2)

			var deployment appsv1.Deployment
			err = r.Get(ctx, types.NamespacedName{Name: topics[piece], Namespace: org.Spec.Consumers.Namespace}, &deployment)
			if err != nil {
				slog.ErrorContext(ctx, "error while getting deployment "+topics[piece], "err", err)
				continue
			}

			if deployment.Status.ReadyReplicas > 0 {
				isSuccess = true
				break
			}
		}

		if !isSuccess {
			slog.ErrorContext(ctx, "retry exaused while getting consumer-"+topics[piece])
			return ctr.Result{}, err
		}
	}

	return ctr.Result{}, nil
}

func (r *OrgReconciler) SetupWithManager(mgr ctr.Manager) error {
	return ctr.NewControllerManagedBy(mgr).
		For(&crdv1.Org{}).
		Named("org").
		Complete(r)
}

func (r *OrgReconciler) CreateIfNotExist(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {

	err := r.Get(ctx, key, obj, opts...)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Create(ctx, obj)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func QueueService(org crdv1.Org) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      org.Spec.Queue.Name,
			Namespace: org.Spec.Queue.Namespace,
			Labels:    map[string]string{"app": "queue"},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: "plaintext", Port: org.Spec.Queue.Ports.Plaintext, TargetPort: intstr.FromInt32(org.Spec.Queue.Ports.Plaintext)},
				{Name: "controller", Port: org.Spec.Queue.Ports.Controller, TargetPort: intstr.FromInt32(org.Spec.Queue.Ports.Controller)},
				{Name: "external", Port: org.Spec.Queue.Ports.External, TargetPort: intstr.FromInt32(org.Spec.Queue.Ports.External)},
			},
			Selector: map[string]string{"app": "queue"},
		},
	}
}

func QueueDeployment(org crdv1.Org) *appsv1.Deployment {
	labels := map[string]string{"app": "queue"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "queue", Namespace: org.Spec.Queue.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: org.Spec.Queue.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Name: "queue", Namespace: org.Spec.Queue.Namespace, Labels: labels},
				Spec: corev1.PodSpec{Containers: []corev1.Container{
					{
						Name:  "kafka",
						Image: "confluentinc/cp-kafka:7.4.0",
						Ports: []corev1.ContainerPort{
							{Name: "plaintext", ContainerPort: org.Spec.Queue.Ports.Plaintext},
							{Name: "controller", ContainerPort: org.Spec.Queue.Ports.Controller},
							{Name: "extneral", ContainerPort: org.Spec.Queue.Ports.External},
						},
						Env: []corev1.EnvVar{
							// Broker identity & cluster config
							{Name: "KAFKA_BROKER_ID", Value: "1"},
							{Name: "KAFKA_NODE_ID", Value: "1"},
							{Name: "CLUSTER_ID", Value: org.Spec.Queue.ClusterId},

							// Listeners & networking
							{Name: "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP", Value: "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT,EXTERNAL:PLAINTEXT"},
							{Name: "KAFKA_ADVERTISED_LISTENERS", Value: fmt.Sprintf(
								"PLAINTEXT://%v.%v.svc.cluster.local:%v,EXTERNAL://localhost:%v",
								org.Spec.Queue.Name, org.Spec.Queue.Namespace, org.Spec.Queue.Ports.Plaintext, org.Spec.Queue.Ports.External,
							)},
							{Name: "KAFKA_LISTENERS", Value: fmt.Sprintf(
								"PLAINTEXT://0.0.0.0:%v,CONTROLLER://0.0.0.0:%v,EXTERNAL://0.0.0.0:%v",
								org.Spec.Queue.Ports.Plaintext, org.Spec.Queue.Ports.Controller, org.Spec.Queue.Ports.External,
							)},

							// Broker & quorum
							{Name: "KAFKA_PROCESS_ROLES", Value: "broker,controller"},
							{Name: "KAFKA_CONTROLLER_LISTENER_NAMES", Value: "CONTROLLER"},
							{Name: "KAFKA_CONTROLLER_QUORUM_VOTERS", Value: fmt.Sprintf(
								"1@%v.%v.svc.cluster.local:9093",
								org.Spec.Queue.Name, org.Spec.Queue.Namespace),
							},
							{Name: "KAFKA_INTER_BROKER_LISTENER_NAME", Value: "PLAINTEXT"},

							// Consumer group behavior
							{Name: "KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS", Value: "0"},

							// Misc
							{Name: "KAFKA_AUTO_CREATE_TOPICS_ENABLE", Value: "true"},
							{Name: "KAFKA_LOG_DIRS", Value: "/tmp/kraft-combined-logs"},
							{Name: "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR", Value: fmt.Sprintf("%v", *org.Spec.Queue.Replicas)},
							{Name: "KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR", Value: "1"},
							{Name: "KAFKA_TRANSACTION_STATE_LOG_MIN_ISR", Value: "1"},
						},
					}},
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}
}

func ConsumerDeployment(org crdv1.Org, topic string) *appsv1.Deployment {

	labels := map[string]string{"consumer": topic}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: topic, Namespace: org.Spec.Consumers.Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Name: topic, Namespace: org.Spec.Consumers.Namespace, Labels: labels},
				Spec: corev1.PodSpec{Containers: []corev1.Container{
					{
						Name:            topic,
						Image:           "backend:latest",
						ImagePullPolicy: "Never",
						Env: []corev1.EnvVar{
							{Name: "BROKER", Value: fmt.Sprintf(
								"%v.%v.svc.cluster.local:%v",
								org.Spec.Queue.Name, org.Spec.Queue.Namespace, org.Spec.Queue.Ports.Plaintext,
							)},
							{Name: "TOPIC_NAME", Value: topic},
							{Name: "GROUP_ID", Value: org.Spec.Consumers.GroupId},
						},
						Args: []string{"sh", "-c", "cd backend/consumer && go run main.go"},
					},
				}},
			},
		},
	}
}
