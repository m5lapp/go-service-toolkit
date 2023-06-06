package config

import (
	"flag"
	"strings"
)

// Cors stores the configuration for CORS (Cross-Origin Resource Sharing).
type Cors struct {
	trustedOrigins []string
}

// Flags parses the flags configured for CORS.
func (c *Cors) Flags() {
	flag.Func(
		"cors-trusted-origins",
		"Trusted CORS origins (space seperated)",
		func(val string) error {
			c.trustedOrigins = strings.Fields(val)
			return nil
		},
	)
}

// Limiter stores the configuration for a rate limiter.
type Limiter struct {
	rps    float64
	burst  int
	active bool
}

// Flags parses the flags configured for a rate limiter. The parameters it
// takes are the default values to use for the rps, burst and active flags
// respectively.
func (l *Limiter) Flags(rps float64, burst int, active bool) {
	flag.Float64Var(&l.rps, "limiter-rps", rps, "Rate limiter max requests per second")
	flag.IntVar(&l.burst, "limiter-burst", burst, "Rate limiter max burst per second")
	flag.BoolVar(&l.active, "limiter-active", active, "Activate rate limiter")
}

// MongoDB stores the configuration for a MongoDB NoSQL database.
type MongoDB struct {
	host       string
	schema     string
	privateKey string
}

// Flags parses the flags for a MongoDB database.
func (m *MongoDB) Flags() {
	flag.StringVar(&m.host, "mongo-host", "", "MongoDB hostname")
	flag.StringVar(&m.schema, "mongo-schema", "", "MongoDB cluster name")
	flag.StringVar(&m.privateKey, "mongo-key", "", "Private key path for MongoDB")
}

// Server stores the configuration for a web application server.
type Server struct {
	addr string
	env  string
}

// Flags parses the flags for a web application server. The parameter is
// for the default server address.
func (s *Server) Flags(addr string) {
	flag.StringVar(&s.addr, "addr", addr, "HTTP address in format: [HOST]:PORT")
	flag.StringVar(&s.env, "env", "development", "Environment (development|staging|production)")
}

// Smtp stores the configuration for an SMTP server connection.
type Smtp struct {
	host     string
	port     int
	username string
	password string
	sender   string
}

// Flags parses the flags for an SMTP server connection.
func (s *Smtp) Flags(host, sender string) {
	flag.StringVar(&s.host, "smtp-host", host, "SMTP host")
	flag.IntVar(&s.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&s.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&s.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&s.sender, "smtp-sender", sender, "SMTP sender, format: Name <email@address.com>")
}

// SqlDB stores the configuration for a SQL database.
type SqlDB struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

// Flags parses the flags for a SQL database. The parameters it takes are
// for the default max open connections, max idle connections and max idle
// connection times respectively.
func (s *SqlDB) Flags(open, idle int, idleTime string) {
	flag.StringVar(&s.dsn, "db-dsn", "", "Database DSN (Data Source Name)")
	flag.IntVar(&s.maxOpenConns, "db-max-open-conns", open, "Database max open connections")
	flag.IntVar(&s.maxIdleConns, "db-max-idle-conns", idle, "Database max idle connections")
	flag.StringVar(&s.maxIdleTime, "db-max-idle-time", idleTime, "Database max connection idle time (time.Duration)")
}
