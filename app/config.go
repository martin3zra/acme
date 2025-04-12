package app

import (
	"os"
	"strconv"
)

type Config struct {
	appName      string
	isProduction bool
	host         string
	port         string
	db           Database
	session      Session
}

func (c *Config) ensureHasBeenSet() {
	if c.host == "" {
		panic("Host should be set as environment variable")
	}
	if c.port == "" {
		panic("App port number should be set as environment variable")
	}

	if c.db.name == "" {
		panic("Database name should be set as environment variable")
	}

	if c.db.username == "" {
		panic("Database user should be set as environment variable")
	}

	if c.db.password == "" {
		panic("Database password should be set as environment variable")
	}

	if c.db.host == "" {
		panic("Database host should be set as environment variable")
	}

	if c.db.port == "" {
		panic("Database port number should be set as environment variable")
	}

}

func LoadConfig() *Config {
	var isProduction bool
	if os.Getenv("APP_ENV") == "prod" {
		isProduction = true
	}

	lifetimeSession, _ := strconv.Atoi(os.Getenv("SESSION_LIFETIME"))

	return &Config{
		appName:      os.Getenv("APP_NAME"),
		isProduction: isProduction,
		host:         os.Getenv("APP_URL"),
		port:         os.Getenv("APP_PORT"),
		db: Database{
			name:     os.Getenv("DB_NAME"),
			username: os.Getenv("DB_USERNAME"),
			password: os.Getenv("DB_PASSWORD"),
			host:     os.Getenv("DB_HOST"),
			port:     os.Getenv("DB_PORT"),
		},
		session: Session{
			lifetime: lifetimeSession,
			cookie:   "acme_session",
			domain:   "https://acme.com",
			secure:   isProduction,
			httpOnly: true,
		},
	}
}

type Database struct {
	name     string
	username string
	password string
	host     string
	port     string
}

type Session struct {
	// Here you may specify the number of minutes that you wish the session
	// to be allowed to remain idle before it expires. If you want them
	// to expire immediately when the browser is closed then you may
	// indicate that via the expire_on_close configuration option.
	lifetime int
	// Here you may change the name of the session cookie that is created by
	// the framework. Typically, you should not need to change this value
	// since doing so does not grant a meaningful security improvement.
	cookie string
	// This value determines the domain and subdomains the session cookie is
	// available to. By default, the cookie will be available to the root
	// domain and all subdomains. Typically, this shouldn't be changed.
	domain string
	// HTTPS Only Cookies
	// By setting this option to true, session cookies will only be sent back
	// to the server if the browser has a HTTPS connection. This will keep
	// the cookie from being sent to you when it can't be done securely.
	secure bool
	// Setting this value to true will prevent JavaScript from accessing the
	// value of the cookie and the cookie will only be accessible through
	// the HTTP protocol. It's unlikely you should disable this option.
	httpOnly bool
}
