package enviroment

import (
	"os"

	"github.com/joho/godotenv"
)

var JwtKey []byte
var SMTPKey string
var SMTPEMAIL string

func SetEnv() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	JwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	SMTPKey = os.Getenv("GOOGLE_SMTP")
	SMTPEMAIL = os.Getenv("SMTP_PASS_EMAIL")
}
