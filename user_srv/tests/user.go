package main

import (
	"context"
	"fmt"
	"log"
	proto "shop_servs/user_srv/proto/user"

	"google.golang.org/grpc"
)

var userClient proto.UserClient
var coon *grpc.ClientConn

func Init() {
	var err error
	coon, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	userClient = proto.NewUserClient(coon)
}

func TestGetUserList() {
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 30,
	})
	if err != nil {
		panic(err)
	}
	for _, user := range rsp.Data {
		fmt.Println(user.Mobile, user.NickName, user.Password)
		checkrsp, err := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			Password:          "admin123",
			EncryptedPassword: user.Password,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(checkrsp.Success)
	}

}

func TestCreateUser() {
	for i := 0; i < 10; i++ {
		rsp, err := userClient.CreateUser(context.Background(), &proto.CreateUserInfo{
			NickName: fmt.Sprintf("user%d", i),
			Mobile:   fmt.Sprintf("1380000000%d", i),
			Password: "admin123",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(rsp.Id)
	}

}

func main() {
	Init()
	//TestCreateUser()
	TestGetUserList()

	coon.Close()

}
