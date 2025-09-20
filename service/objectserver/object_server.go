package objectserver

import (
	"myGinServer/config"
	"myGinServer/pkg/minio"
)

type ObjectServer struct {
	Cli minio.ObjectStoreClient
}

func NewObjectServer(cfg *config.ObjectServerConfig) *ObjectServer {
	minioClient, err := minio.NewMinioClient(cfg.Endpoint, cfg.AccessID, cfg.AccessSecret, cfg.BucketName, cfg.RootPath)
	if err != nil {
		panic(err)
	}
	return &ObjectServer{
		Cli: minioClient,
	}
}
