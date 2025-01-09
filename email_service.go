package main

import (
	"email/Inits"
	"email/global"
	"email/router"
	"email/service"
	"email/service/aws"
	"email/service/dovecot"
)

// 初始化
func initEnv() {
	//初始化配置
	Inits.InitConfig()
	//初始化log
	Inits.InitLogger()
	//初始化数据库链接
	Inits.InitPsql()
}

func main() {
	initEnv()

	//err := dao.AddDomain("mountex.online", "admin@mountex.online")
	//if err != nil {
	//	fmt.Printf("添加域名失败: %v\n", err)
	//}
	//

	//err := dao.AddAccount("mountex.online", "Test1", "test1", "Devtest77!")
	//if err != nil {
	//	fmt.Printf("添加账户失败: %v\n", err)
	//}

	global.Log.Info("服务开始启动...")
	if global.Config.System.Env {
		go dovecot.DovecotStatusMonitorInit()
		go service.SmtpServerInit()
	}
	go router.InitRouter()
	aws.ProcessSQSEmailMessages()

}
