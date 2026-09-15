package main

import (
	"context"
	"myGinServer/config"
	"myGinServer/controller"
	"myGinServer/internal/store"
	"myGinServer/pkg/metrics"
	"myGinServer/router"
)

const (
	prometheusMonitor = false
)

func main() {
	configYaml := config.InitConfig("./config.yaml")

	db, err := store.NewDatabase(&configYaml.DBConfig)
	if err != nil {
		panic(err)
	}

	userController := controller.NewUserController(db, configYaml)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// chatController := controller.NewChatController(configYaml)
	dispatch := controller.NewDispatchController(ctx, configYaml)
	metricsCollector := metrics.NewMetrics()

	r := router.NewRouter(userController,
		// chatController,
		dispatch,
		metricsCollector, db)

	if prometheusMonitor {
		go func() {
			router.NewPrometheusMetricsEngine().Run(":18081")
		}()
	}

	err = r.Start(":8081")
	if err != nil {
		panic(err)
	}

}
