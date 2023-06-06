package config

import (
	"flag"
	"strings"
)

// Cors stores the configuration for CORS (Cross-Origin Resource Sharing).
type Cors struct {
	TrustedOrigins []string
}

// Flags parses the flags configured for CORS.
func (c *Cors) Flags() {
	flag.Func(
		"cors-trusted-origins",
		"Trusted CORS origins (space seperated)",
		func(val string) error {
			c.TrustedOrigins = strings.Fields(val)
			return nil
		},
	)
}

// Limiter stores the configuration for a rate limiter.
type Limiter struct {
	RPS    float64
	Burst  int
	Active bool
}

// Flags parses the flags configured for a rate limiter. The parameters it
// takes are the default values to use for the rps, burst and active flags
// respectively.
func (l *Limiter) Flags(rps float64, burst int, active bool) {
	flag.Float64Var(&l.RPS, "limiter-rps", rps, "Rate limiter max requests per second")
	flag.IntVar(&l.Burst, "limiter-burst", burst, "Rate limiter max burst per second")
	flag.BoolVar(&l.Active, "limiter-active", active, "Activate rate limiter")
}

// MongoDB stores the configuration for a MongoDB NoSQL database.
type MongoDB struct {
	Host       string
	Schema     string
	PrivateKey string
}

// Flags parses the flags for a MongoDB database.
func (m *MongoDB) Flags() {
	flag.StringVar(&m.Host, "mongo-host", "", "MongoDB hostname")
	flag.StringVar(&m.Schema, "mongo-schema", "", "MongoDB cluster name")
	flag.StringVar(&m.PrivateKey, "mongo-key", "", "Private key path for MongoDB")
}

// Server stores the configuration for a web application server.
type Server struct {
	Addr string
	Env  string
}

// Flags parses the flags for a web application server. The parameter is
// for the default server address.
func (s *Server) Flags(addr string) {
	flag.StringVar(&s.Addr, "addr", addr, "HTTP address in format: [HOST]:PORT")
	flag.StringVar(&s.Env, "env", "development", "Environment (development|staging|production)")
}

// Smtp stores the configuration for an SMTP server connection.
type Smtp struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

// Flags parses the flags for an SMTP server connection.
func (s *Smtp) Flags(host, sender string) {
	flag.StringVar(&s.Host, "smtp-host", host, "SMTP host")
	flag.IntVar(&s.Port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&s.Username, "smtp-username", "", "SMTP username")
	flag.StringVar(&s.Password, "smtp-password", "", "SMTP password")
	flag.StringVar(&s.Sender, "smtp-sender", sender, "SMTP sender, format: Name <email@address.com>")
}

// SqlDB stores the configuration for a SQL database.
type SqlDB struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

// Flags parses the flags for a SQL database. The parameters it takes are
// for the default max open connections, max idle connections and max idle
// connection times respectively.
func (s *SqlDB) Flags(open, idle int, idleTime string) {
	flag.StringVar(&s.DSN, "db-dsn", "", "Database DSN (Data Source Name)")
	flag.IntVar(&s.MaxOpenConns, "db-max-open-conns", open, "Database max open connections")
	flag.IntVar(&s.MaxIdleConns, "db-max-idle-conns", idle, "Database max idle connections")
	flag.StringVar(&s.MaxIdleTime, "db-max-idle-time", idleTime, "Database max connection idle time (time.Duration)")
}
