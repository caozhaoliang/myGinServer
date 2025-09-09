package main

import (
	"myGinServer/config"
	"myGinServer/controller"
	"myGinServer/internal/store"
	"myGinServer/router"
	"myGinServer/service/userserver"
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
	userServer := userserver.NewUserServer(db)

	userController := controller.NewUserController(userServer)

	r := router.NewRouter(userController, db)

	if prometheusMonitor {
		go func() {
			router.NewPromethusMetricsEngine().Run(":18081")
		}()
	}

	err = r.Start(":8080")
	if err != nil {
		panic(err)
	}

}
