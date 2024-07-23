package controller

import (
	"context"

	appv1 "k8s.io/api/apps/v1"

	"github.com/steeling/controller-runtime-exercise/pkg/api"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	client  client.Client
	manager ctrl.Manager
}

func New() (*Controller, error) {
	manager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		return nil, err
	}
	controller := &Controller{
		client:  manager.GetClient(),
		manager: manager,
	}

	err = ctrl.
		NewControllerManagedBy(manager). // Create the Controller
		For(&api.MyApp{}).               // MyApp is the Application API
		Owns(&appv1.Deployment{}).       // MyApp owns Deployments created by it
		Complete(controller)
	if err != nil {
		return nil, err
	}
	return controller, nil
}

func (c *Controller) Start(ctx context.Context) error {
	return c.manager.Start(ctx)
}

func (c *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: implement
	return ctrl.Result{}, nil
}
