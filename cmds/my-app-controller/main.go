package main

import (
	"context"

	"github.com/steeling/controller-runtime-exercise/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	log.SetLogger(zap.New(zap.UseDevMode(true)))
	ctx := context.Background()
	// Create a new controller
	c, err := controller.New()
	check(err)

	// Start the controller
	c.Start(ctx)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
