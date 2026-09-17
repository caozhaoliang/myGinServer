package main

import (
	"context"
	"fmt"
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

	// 初始化默认租户并把 admin 用户绑定为 owner（幂等）。
	if err := db.EnsureDefaultTenant(context.Background()); err != nil {
		fmt.Printf("初始化默认租户失败: %v\n", err)
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
