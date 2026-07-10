package app

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/martin3zra/forge/mailer"
)

type ConfigKey struct{}
type Config struct {
	appName      string
	isProduction bool
	host         string
	port         string
	db           Database
	session      Session
	secretKey    []byte
	mail         Mail
	sse          SSE
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

	if string(c.secretKey) == "" {
		panic("APP_KEY should be set as environment variable")
	}

	if c.sse.port == "" {
		panic("SSE port number should be set as environment variable")
	}

	if c.sse.url == "" {
		panic("SSE URL should be set as environment variable")
	}

}

func LoadConfig() *Config {
	var isProduction bool
	if os.Getenv("APP_ENV") == "prod" {
		isProduction = true
	}

	lifetimeSession, _ := strconv.Atoi(os.Getenv("SESSION_LIFETIME"))

	ssePort := os.Getenv("SSE_PORT")
	if ssePort == "" {
		ssePort = defaultSSEPort
	}

	encoded := os.Getenv("APP_KEY")
	if encoded == "" {
		log.Fatal("APP_KEY environment variable is required")
	}

	encoded = strings.TrimPrefix(encoded, "base64:")

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Fatalf("Failed to decode APP_KEY: %v", err)
	}

	if len(key) != 64 {
		log.Fatalf("APP_KEY must be 64 bytes when decoded, got %d bytes", len(key))
	}

	return &Config{
		appName:      os.Getenv("APP_NAME"),
		secretKey:    key,
		isProduction: isProduction,
		host:         os.Getenv("APP_URL"),
		port:         os.Getenv("APP_PORT"),
		db: Database{
			name:     os.Getenv("DB_NAME"),
			username: os.Getenv("DB_USERNAME"),
			password: os.Getenv("DB_PASSWORD"),
			host:     os.Getenv("DB_HOST"),
			port:     os.Getenv("DB_PORT"),
			sslmode:  os.Getenv("DB_SSLMODE"),
		},
		session: Session{
			lifetime: lifetimeSession,
			cookie:   "acme_session",
			domain:   "https://acme.com",
			secure:   isProduction,
			httpOnly: true,
		},
		mail: Mail{
			Driver:      MailDriver(os.Getenv("MAIL_DRIVER")),
			Host:        os.Getenv("MAIL_HOST"),
			Port:        os.Getenv("MAIL_PORT"),
			FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
			FromName:    os.Getenv("MAIL_FROM_NAME"),
			Username:    os.Getenv("MAIL_USERNAME"),
			Password:    os.Getenv("MAIL_PASSWORD"),
			Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		},
		sse: SSE{
			port: ssePort,
			url:  resolveSSEURL(os.Getenv("SSE_URL"), os.Getenv("APP_URL"), ssePort),
		},
	}
}

const defaultSSEPort = "8090"

// SSE holds the event-stream listener's port and the absolute base URL the
// browser dials. They differ whenever a reverse proxy terminates the stream on
// a different host, scheme, or port than the one the process binds.
type SSE struct {
	port string
	url  string
}

// resolveSSEURL prefers an explicit SSE_URL, which is what a deployment behind
// a proxy or TLS terminator needs. Otherwise it reuses APP_URL's scheme and
// host with the SSE port swapped in, so a stock local setup needs no extra
// configuration. The value is served to the browser as a shared Inertia prop
// rather than baked into the bundle, so one build runs in every environment.
func resolveSSEURL(explicit, appURL, ssePort string) string {
	if explicit != "" {
		return strings.TrimSuffix(explicit, "/")
	}

	parsed, err := url.Parse(appURL)
	if err != nil || parsed.Hostname() == "" {
		return "http://localhost:" + ssePort
	}

	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(parsed.Hostname(), ssePort))
}

type Database struct {
	name     string
	username string
	password string
	host     string
	port     string
	sslmode  string
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

type MailDriver string

const (
	STMP MailDriver = "smtp"
	API  MailDriver = "api"
)

type Mail struct {
	Driver      MailDriver
	Host        string
	Port        string
	FromAddress string
	FromName    string
	Username    string
	Password    string
	Encryption  string
}

func (m Mail) asMailConfig() mailer.Config {
	return mailer.Config{
		Driver:      mailer.MailDriver(m.Driver),
		Host:        m.Host,
		Port:        m.Port,
		FromAddress: m.FromAddress,
		FromName:    m.FromName,
		Username:    m.Username,
		Password:    m.Password,
		Encryption:  m.Encryption,
	}
}
