package nosql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
	Id     bson.ObjectID `bson:"_id"`
	UserId string        `bson:"user_id"`
	Name   string        `bson:"name"`
	Email  string        `bson:"email"`
	Age    int           `bson:"age"`
}

func TestConnect(t *testing.T) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	//SetAuth(options.Credential{
	//			Username:   "username",      // 用户名
	//			Password:   "password",      // 密码
	//			AuthSource: "admin",         // 认证数据库
	//		}).
	//		SetConnectTimeout(10 * time.Second).    // 连接超时时间
	//		SetMaxPoolSize(100)                     // 连接池大小
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	coll := client.Database("myapp").Collection("movies")
	title := "Back to the Future"
	var result bson.M
	err = coll.FindOne(context.TODO(), bson.D{{"title", title}}).
		Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		fmt.Printf("No document was found with the title %s\n", title)
		return
	}
	if err != nil {
		panic(err)
	}
	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", jsonData)
}
