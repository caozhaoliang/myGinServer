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
	Id     bson.ObjectID `bson:"_id,omitempty"`
	UserId string        `bson:"user_id"`
	Name   string        `bson:"name"`
	Email  string        `bson:"email"`
	Age    int           `bson:"age"`
}

func createClient() *mongo.Client {
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
	return client
}
func disconnect(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func TestFindOne(t *testing.T) {
	client := createClient()
	var err error

	coll := client.Database("myapp").Collection("user")
	var result bson.M
	err = coll.FindOne(context.TODO(), bson.D{{"name", "WangWu"}}).
		Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		fmt.Printf("No document was found ")
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

func TestDelete(t *testing.T) {
	client := createClient()
	coll := client.Database("myapp").Collection("user")
	deleteResult, err := coll.DeleteOne(context.TODO(), bson.D{{"name", "ZhangSan"}})
	fmt.Println(deleteResult, err)
}

func TestInsertOne(t *testing.T) {
	user := User{
		UserId: "myUser10001",
		Name:   "ZhangSan",
		Email:  "ZhangSan@gmail.com",
		Age:    18,
	}
	client := createClient()
	collection := client.Database("myapp").Collection("user")

	insertOne, err := collection.InsertOne(context.Background(), user)
	fmt.Println(insertOne, err)
}

func TestInsertMany(t *testing.T) {
	client := createClient()
	collection := client.Database("myapp").Collection("user")
	users := []interface{}{
		User{
			UserId: "myUser10001",
			Name:   "ZhangSan",
			Email:  "ZhangSan@gmail.com",
			Age:    18,
		},
		User{
			UserId: "myUser10002",
			Name:   "LiSi",
			Email:  "LiSi@gmail.com",
			Age:    19,
		},
		User{
			UserId: "myUser10003",
			Name:   "WangWu",
		},
	}
	insertMany, err := collection.InsertMany(context.Background(), users)
	fmt.Println(insertMany, err)
}

func TestUpdateOne(t *testing.T) {
	client := createClient()
	collection := client.Database("myapp").Collection("user")
	updateResult, err := collection.UpdateOne(context.Background(),
		bson.D{{"name", "WangWu"}},
		bson.D{{"$set", bson.D{{"email", "WangWu@gmail.com"}, {"age", 20}}}})
	fmt.Println(updateResult, err)
}
