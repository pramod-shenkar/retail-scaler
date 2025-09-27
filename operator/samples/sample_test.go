package samples

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestOperator(t *testing.T) {

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	s := runtime.NewScheme()
	scheme.AddToScheme(s)

	kubeconfig, err := config.GetConfig()
	require.NoError(t, err)

	manager, err := ctrl.NewManager(
		kubeconfig, manager.Options{
			Scheme: s,
			Logger: logr.FromSlogHandler(tint.NewHandler(os.Stderr, nil))},
	)
	require.NoError(t, err)

	reconciler := NewCustomReconciler(manager)
	reconciler.SetupWithManager(manager)

	manager.Start(t.Context())

}

type CustomReconciler struct {
	client.Client                 // read/write k8s resources. Its using api calls underlaying
	Scheme        *runtime.Scheme // it will maps go structs to k8s api calls ie appsv1.Deployment{} should call /api/apps/v1/deployment api
}

func NewCustomReconciler(manager ctrl.Manager) *CustomReconciler {
	return &CustomReconciler{
		Client: manager.GetClient(),
		Scheme: manager.GetScheme(),
	}
}

func (r CustomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var mydeployment = appsv1.Deployment{}
	var opts = []client.GetOption{}
	err := r.Get(ctx, types.NamespacedName{Name: "indent-consumer", Namespace: "default"}, &mydeployment, opts...)
	if err != nil {
		slog.ErrorContext(ctx, "error while", "err", err.Error())
		return ctrl.Result{}, err
	}

	if *mydeployment.Spec.Replicas < 4 {
		mydeployment.Spec.Replicas = pointer.Int32(4)
		err = r.Update(ctx, &mydeployment)
		if err != nil {
			slog.ErrorContext(ctx, "error while", "err", err.Error())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r CustomReconciler) SetupWithManager(manager ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(manager).
		For(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Watches(&corev1.Secret{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
