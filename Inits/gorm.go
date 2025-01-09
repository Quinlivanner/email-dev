package Inits

import (
	"email/global"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	gorm.Model // 嵌入gorm.Model结构体
	Name       string
	Age        int
}

func InitPsql() {
	if global.Config.Psql.Host == "" || global.Config.Psql.Port == 0 || global.Config.Psql.User == "" || global.Config.Psql.Database == "" {
		panic("Psql 数据库配置不正确，请检查 settings.yaml 文件")
	}
	dsn := global.Config.Psql.DSN()
	var psqlLogger logger.Interface
	if global.Config.Psql.LogLevel == "dev" {
		psqlLogger = logger.Default.LogMode(logger.Info)
	} else {
		psqlLogger = logger.Default.LogMode(logger.Error)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: psqlLogger,
	})
	if err != nil {
		panic("Psql数据库连接失败: " + err.Error())
	}
	sqlDb, err := db.DB()
	if err != nil {
		panic("Psql 实例获取失败: " + err.Error())
	}
	sqlDb.SetMaxIdleConns(global.Config.Psql.MaxIdleConns)
	sqlDb.SetMaxOpenConns(global.Config.Psql.MaxOpenConns)
	sqlDb.SetConnMaxLifetime(time.Hour * time.Duration(global.Config.Psql.ConnMaxLifetime))
	sqlDb.SetConnMaxIdleTime(time.Minute * time.Duration(global.Config.Psql.ConnMaxLifetime))
	global.PsqlDB = db
	global.SqlDB = sqlDb
	global.Log.Info("Psql数据库连接成功...")
}
