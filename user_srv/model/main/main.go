package main

//测试
import (
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"shop_servs/user_srv/model"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// 生成md5密码
func main() {
	dsn := "root:123456@tcp(192.168.88.129:3306)/shop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)
	//连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	options := &password.Options{
		SaltLen:      16,
		Iterations:   100,
		KeyLen:       32,
		HashFunction: sha512.New,
	}
	salt, encodePwd := password.Encode("admin123", options)
	newpassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodePwd)

	for i := 0; i < 10; i++ {
		user := model.User{
			NickName: fmt.Sprintf("user%d", i),
			Mobile:   fmt.Sprintf("188%08d", i),
			Password: newpassword,
		}
		db.Save(&user)
	}
}
