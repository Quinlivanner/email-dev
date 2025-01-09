package service

import (
	"email/controller/response"
	"email/dao"
	"email/global"
	"email/models"
	"email/utils"
	"errors"
	"time"
)

// LoginProcess 登录
func LoginProcess(l models.LoginRequest) (*models.AccessToken, error) {
	// 验证密码
	s, e := dao.ValidateAccount(l.Email, l.Password)
	if e != nil {
		return nil, e
	}
	expirationTime := time.Now().Add(time.Duration(global.Config.Jwt.ExpiredTime) * time.Hour * 24)
	token, err := utils.GenerateJwtToken(s, expirationTime)
	if err != nil {
		//返回错误
		return nil, errors.New(response.GenerateJwtTokenFailedCode.ErrMessage)
	}
	//返回成功
	return &models.AccessToken{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   expirationTime.Unix()},
		nil

}
