package setup

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	g "wp_template_display/internal/global"
	"wp_template_display/internal/models"

	"github.com/gookit/config/v2"
	"golang.org/x/crypto/ssh"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 创建SSH隧道
func createSSHTunnel() (*ssh.Client, net.Listener, error) {
	host := os.Getenv("MYSQL_SSH_TUNNEL_HOST")
	port := os.Getenv("MYSQL_SSH_TUNNEL_PORT")
	user := os.Getenv("MYSQL_SSH_TUNNEL_USER")
	password := os.Getenv("MYSQL_SSH_TUNNEL_PASSWORD")

	// 配置SSH客户端
	sshClientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	// 连接SSH服务器
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), sshClientConfig)
	if err != nil {
		panic(err)
	}

	// 在本地启动监听器
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	// 启动转发
	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			dbHost := config.String("Database.Host", "127.0.0.1")
			dbPort := config.Int("Database.Port", 3306)
			remoteConn, err := sshClient.Dial("tcp", fmt.Sprintf("%s:%d", dbHost, dbPort))
			if err != nil {
				panic(err)
			}

			// 转发数据
			go func() {
				defer localConn.Close()
				defer remoteConn.Close()

				go func() {
					_, _ = io.Copy(localConn, remoteConn)
				}()
				_, _ = io.Copy(remoteConn, localConn)
			}()
		}
	}()

	return sshClient, listener, nil
}

func SetupDatabase() {
	driver := config.String("Database.Driver", "none")
	if driver == "none" {
		panic("The database driver is not configured.")
	} else if driver == "mysql" {
		var dsn string
		MYSQL_SSH_TUNNEL_HOST := os.Getenv("MYSQL_SSH_TUNNEL_HOST")

		if MYSQL_SSH_TUNNEL_HOST == "" {
			// 为空则不经过隧道 直接获取
			dsn = fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				config.String("Database.Username", "root"),
				config.String("Database.Password", "123456"),
				config.String("Database.Host", "127.0.0.1"),
				config.Int("Database.Port", 3306),
				config.String("Database.Database", "wp_template_display"),
			)
		} else {
			// 创建SSH隧道
			_, listener, err := createSSHTunnel()
			if err != nil {
				panic(err)
			}

			// 获取本地监听的地址
			localAddr := listener.Addr().(*net.TCPAddr)

			// 拼接DSN
			dsn = fmt.Sprintf(
				"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				config.String("Database.Username", "root"),
				config.String("Database.Password", "123456"),
				"127.0.0.1",
				localAddr.Port,
				config.String("Database.Database", "wp_template_display"),
			)
		}

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic(err)
		}

		g.SetDatabase(db)
	} else if driver == "sqlite" {
		db, err := gorm.Open(sqlite.Open(config.String("Database.Dbpath", "./runtime/database.db")), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		g.SetDatabase(db)
	} else {
		panic("The database driver is not supported.")
	}

	models.InitModel()
}
