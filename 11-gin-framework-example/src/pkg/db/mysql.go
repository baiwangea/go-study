package db

import (
	"fmt"
	"gin-framework-example/src/pkg/util"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Init() {
	var err error
	dbCfg := util.Conf.DB

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.DBName,
		dbCfg.Charset,
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbCfg.Prefix,
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 数据表已通过迁移工具创建，此处不再执行 GORM 的自动迁移 (AutoMigrate)。
	// 在生产环境中，推荐使用专业的数据库迁移工具（如 migrate, flyway）来管理表结构变更，
	// 以避免潜在的数据丢失风险。

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
}
