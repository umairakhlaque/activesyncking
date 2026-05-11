package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration for all SyncGuard services.
// Values are loaded from a YAML file and overridden by environment variables
// with the prefix SYNCGUARD_ (e.g. SYNCGUARD_DB_HOST → db.host).
type Config struct {
	DB         DBConfig         `mapstructure:"db"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Gateway    GatewayConfig    `mapstructure:"gateway"`
	Admin      AdminConfig      `mapstructure:"admin"`
	Email      EmailConfig      `mapstructure:"email"`
	LDAP       LDAPConfig       `mapstructure:"ldap"`
	Audit      AuditConfig      `mapstructure:"audit"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

type DBConfig struct {
	// DatabaseURL takes full precedence when set (e.g. Neon connection string).
	// Format: postgres://user:pass@host/dbname?sslmode=require
	DatabaseURL string `mapstructure:"database_url"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Name        string `mapstructure:"name"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	SSLMode     string `mapstructure:"sslmode"`
	MaxConns    int    `mapstructure:"max_conns"`
	MinConns    int    `mapstructure:"min_conns"`
}

func (d DBConfig) DSN() string {
	if d.DatabaseURL != "" {
		// Append pool sizing if not already present
		sep := "?"
		if strings.Contains(d.DatabaseURL, "?") {
			sep = "&"
		}
		return fmt.Sprintf("%s%spool_max_conns=%d&pool_min_conns=%d",
			d.DatabaseURL, sep, d.MaxConns, d.MinConns)
	}
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode, d.MaxConns, d.MinConns,
	)
}

type RedisConfig struct {
	// URL takes full precedence when set (e.g. Upstash rediss:// connection string).
	URL      string `mapstructure:"url"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	Listen         string        `mapstructure:"listen"`
	JWTSecret      string        `mapstructure:"jwt_secret"`
	JWTExpiry      time.Duration `mapstructure:"jwt_expiry"`
	ChallengeTTL   time.Duration `mapstructure:"challenge_ttl"`
	MaxAttempts    int           `mapstructure:"max_attempts"`
	DeviceTrustTTL time.Duration `mapstructure:"device_trust_ttl"`
}

type GatewayConfig struct {
	Listen         string        `mapstructure:"listen"`
	TLSCert        string        `mapstructure:"tls_cert"`
	TLSKey         string        `mapstructure:"tls_key"`
	ExchangeURL    string        `mapstructure:"exchange_url"`
	AuthServiceURL string        `mapstructure:"auth_service_url"`
	DialTimeout    time.Duration `mapstructure:"dial_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	FailClosed     bool          `mapstructure:"fail_closed"`
}

type AdminConfig struct {
	Listen      string   `mapstructure:"listen"`
	CORSOrigins []string `mapstructure:"cors_origins"`
	APIKey      string   `mapstructure:"api_key"`
}

type EmailConfig struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
	From         string `mapstructure:"from"`
	FromName     string `mapstructure:"from_name"`
}

type LDAPConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	URL          string `mapstructure:"url"`
	BindDN       string `mapstructure:"bind_dn"`
	BindPassword string `mapstructure:"bind_password"`
	BaseDN       string `mapstructure:"base_dn"`
	UserFilter   string `mapstructure:"user_filter"`
	StartTLS     bool   `mapstructure:"start_tls"`
}

type AuditConfig struct {
	SyslogEnabled bool   `mapstructure:"syslog_enabled"`
	SyslogAddr    string `mapstructure:"syslog_addr"`
	SyslogProto   string `mapstructure:"syslog_proto"`
}

type EncryptionConfig struct {
	// 32-byte AES-256 key, hex encoded
	Key string `mapstructure:"key"`
}

func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.name", "syncguard")
	v.SetDefault("db.user", "syncguard")
	v.SetDefault("db.password", "changeme")
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.max_conns", 20)
	v.SetDefault("db.min_conns", 0)

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)

	v.SetDefault("auth.listen", ":8080")
	v.SetDefault("auth.jwt_expiry", "8h")
	v.SetDefault("auth.challenge_ttl", "5m")
	v.SetDefault("auth.max_attempts", 5)
	v.SetDefault("auth.device_trust_ttl", "720h")

	v.SetDefault("gateway.listen", ":8443")
	v.SetDefault("gateway.dial_timeout", "10s")
	v.SetDefault("gateway.idle_timeout", "90s")
	v.SetDefault("gateway.fail_closed", true)

	v.SetDefault("admin.listen", ":8090")

	v.SetDefault("email.smtp_port", 587)
	v.SetDefault("email.from_name", "SyncGuard MFA")

	v.SetDefault("ldap.user_filter", "(sAMAccountName=%s)")
	v.SetDefault("ldap.start_tls", true)

	v.SetDefault("audit.syslog_proto", "tcp")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("SYNCGUARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return &cfg, nil
}

// ValidateAuth checks that fields required by the auth and crypto subsystems
// are present. Call this in binaries that use JWT or encryption (authsvc,
// adminsvc) but NOT in utilities like the migration runner.
func (c *Config) ValidateAuth() error {
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret must not be empty")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("auth.jwt_secret must be at least 32 characters")
	}
	if c.Encryption.Key == "" {
		return fmt.Errorf("encryption.key must not be empty")
	}
	return nil
}
