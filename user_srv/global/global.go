package global

import (
	"shop_servs/user_srv/config"

	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	ServerConfig *config.ServerConfig
)
