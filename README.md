# Mountex Email 项目部署文档

## 1. 项目简介
基于 Go 语言开发的自托管邮件系统，集成了 AWS 服务，提供完整的邮件收发、存储和管理功能。

## 2. 环境准备
### 2.1 依赖安装
将代码clone到本地后，进入文件夹使用以下命令安装Package.
```
go mod tidy
```
如你的go版本与Mod文件中不一致，请修改Mod文件
```
go 1.22.0 => go 1.xx.x
```

## 3. 项目配置
### 3.1 配置文件说明

API 配置
```
  email_count_per_page: 0               # 每页显示的邮件数量
  short_url_domain: "example.com"       # 短链接域名
  short_url_code_length: 17             # 短链接标识符长度
```
AWS 服务配置
```
  s3_bucket: ""                         # S3存储桶名称
  sqs_url:   ""                         # SQS队列URL
  max_file_size: 20                     # 附件最大大小(MB)
  file_expire_time: 7                   # 文件过期时间(天)
```
PostgreSQL 数据库配置
```
  host: localhost                       # 数据库主机地址
  port: 5432                            # 数据库端口
  user: user                            # 数据库用户名
  password: 'password'                  # 数据库密码
  database: postgres                    # 数据库名称
  max_idle_conns: 10                    # 最大空闲连接数
  max_open_conns: 20                    # 最大打开连接数
  conn_max_lifetime: 1                  # 连接最大生命周期(小时)
  conn_max_idle_time: 1                 # 连接最大空闲时间(小时)
  log_level: error                      # 日志级别
```
JWT 认证配置
```
  expired_time: 300                     # JWT过期时间(秒)
  secret_key: "secret_key"              # JWT密钥
```

日志配置
```
  level: debug                          # 日志级别
  prefix: 'Email'                       # 日志前缀 
  directory: /var/log                   # 日志目录
  show_line: true                       # 显示行号
  show_file_name: false                 # 显示文件名
```
数据库表名配置
```
  domains:         domains              # 域名表
  email_accounts:  email_accounts       # 邮箱账户表
  attachments:     attachments          # 附件表
```
机器人配置
```
  bot_token: ""                         # 机器人Token(可接入第三方平台进行消息预警)
```
AI接口配置
```
  ai_chat_api:                          # AI聊天API
  ai_chat_key: sk-***                   # AI API密钥
  ai_model_name: deepseek-chat          # AI模型名称
  ai_prompt: ""                         # AI提示词
  ```
### 3.2 AWS凭证配置
请提前下载 `AWS CLI`并配置凭证，配置方法请参考 [AWS CLI 配置](https://docs.aws.amazon.com/zh_cn/cli/latest/userguide/cli-configure-files.html)，运行时，程序会自动从本地环境变量中读取凭证，无需单独添加。

## 4. 部署步骤
### 4.1 获取代码
```
git clone https://github.com/Quinlivanner/Email.git
```
### 4.2 安装依赖
```
go mod init email
go mod tidy
```
### 4.3 确认配置文件
在开始运行前，请确保配置文件正确，确保数据库连接正常。配置文件摆放位置为./settings.yaml

### 4.4 运行项目
项目可以直接通过 `go run email_service.go` 运行，也可以编译成对应平台的可执行文件。

>直接运行
```
cd /Email
go run main.go
```
>编译为对应平台可执行文件运行
```
cd /Email

# 编译Linux版本
GOOS=linux GOARCH=amd64 go build -o email_linux

# 编译Windows版本
GOOS=windows GOARCH=amd64 go build -o email.exe

# 编译MacOS版本
GOOS=darwin GOARCH=amd64 go build -o email_mac

编译后运行可以直接运行或后台运行

# 直接运行
./email_linux

# 后台运行
nohup ./email_linux & //日志会输出到nohup.out文件中
```

## 5. 注意事项

### 5.1 附件大小限制
- 单个附件大小限制为20MB
- 附件保存在AWS S3中,保存期限为7天，超过7天需重新请求生成新的下载链接。

### 5.2 API限制
- 每页邮件数量限制为20封
- 需要分页获取邮件列表



## 6.从头开始
### 6.1 创建Domain表

```sql
CREATE TABLE domains
(
    id          SERIAL PRIMARY KEY,               -- 域名唯一标识符
    domain_name VARCHAR(255) UNIQUE     NOT NULL, -- 域名名称
    admin_email VARCHAR(255)            NOT NULL, -- 管理员邮箱
    created_at  TIMESTAMP DEFAULT NOW() NOT NULL-- 记录创建时间
);
```

### 6.2 创建Email Account表

```sql
CREATE TABLE email_accounts
(
    id             SERIAL PRIMARY KEY,                                  -- 邮箱账号唯一标识符
    domain_id      INT                 NOT NULL,                        -- 关联的域名
    domain_name    VARCHAR(255)        NOT NULL,                        -- 域名名称
    email_address  VARCHAR(255) UNIQUE NOT NULL,                        -- 邮箱地址
    password_hash  VARCHAR(255)        NOT NULL,                        -- 哈希后的密码
    jwt_token_hash TEXT,                                                -- jwt token
    user_name      VARCHAR(255)        NOT NULL,                        -- 用户名
    status         VARCHAR(50)         NOT NULL DEFAULT 'active',       -- 账号状态，默认值为 'active'
    storage_used   BIGINT              NOT NULL DEFAULT 0,              -- 用户使用的存储空间（字节）
    created_at     TIMESTAMP                    DEFAULT NOW() NOT NULL, -- 账号创建时间

    -- 外键约束定义
    FOREIGN KEY (domain_id)
        REFERENCES domains (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);
```

### 6.3 创建Attachment表

```sql
CREATE TABLE attachments
(
    id                SERIAL PRIMARY KEY,                                                 -- 附件唯一标识符，自增
    file_hash         CHAR(64) UNIQUE            NOT NULL,                                -- 文件哈希值，唯一索引
    file_name         VARCHAR(255)               NOT NULL,                                -- 文件名
    file_type         VARCHAR(255)               NOT NULL,                                -- 文件类型
    file_size         BIGINT                     NOT NULL,                                -- 文件大小（字节）
    s3_from_email_key VARCHAR(512)               NOT NULL,                                -- S3 存储键
    short_url_code    VARCHAR(255)               NOT NULL,                                -- 短链接
    download_url      VARCHAR(1024)              NOT NULL,                                -- 下载URL
    s3_storage_path   VARCHAR(512) DEFAULT 'N/A' NOT NULL,                                -- 存储路径
    expire_time       TIMESTAMP    DEFAULT CURRENT_TIMESTAMP + INTERVAL '6 days' NOT NULL -- 过期时间
);
```

### 6.4 添加域名

```go

	err := dao.AddDomain("example.com", "admin@example.com")
	if err != nil {
		fmt.Printf("添加域名失败: %v\n", err)
	}

```

### 6.5 添加账户
    
```go
	err := dao.AddAccount("example.com", "Test", "test", "Mountext666!")
	if err != nil {
		fmt.Printf("添加账户失败: %v\n", err)
	}
```

 - 注意事项：
    - 1. 添加账户时，密码会被哈希后存储，所以请确保密码正确。
    - 2. 添加域名或账户只需要将`email_service.go`中的`dao.AddDomain`或`dao.AddAccount`函数调用放到`main`函数中即可。我已经把它们注释了，你取消对应的注释并且填上数据即可。
    - 3. `Settings.yaml`配置文件不用做修改，拉取最新代码创建好表后直接运行，目前从SNS订阅通知，不同的SQS不会互相干扰。