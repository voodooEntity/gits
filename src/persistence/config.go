package persistence

import "fmt"

// AdapterKind identifies the backend adapter selected by PersistenceConfig.
type AdapterKind string

const (
	AdapterKindNone       AdapterKind = "none"
	AdapterKindFilesystem AdapterKind = "filesystem"
	AdapterKindPostgreSQL AdapterKind = "postgresql"
)

// FilesystemConfig configures the filesystem persistence backend.
type FilesystemConfig struct {
	Directory string
}

// PostgresConfig configures the PostgreSQL persistence backend.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	DSN      string
}

// PersistenceConfig selects an optional persistence adapter for a storage instance.
type PersistenceConfig struct {
	Enabled           bool
	Kind              AdapterKind
	ChannelBufferSize int
	Filesystem        *FilesystemConfig
	PostgreSQL        *PostgresConfig
}

// IsNone reports whether this configuration represents no persistence backend.
func (c PersistenceConfig) IsNone() bool {
	return !c.Enabled || c.Kind == "" || c.Kind == AdapterKindNone
}

// Validate checks adapter-specific settings required by known built-in adapters.
// Unknown kinds are allowed so registered custom adapters can perform their own validation.
func (c PersistenceConfig) Validate() error {
	if c.ChannelBufferSize < 0 {
		return fmt.Errorf("persistence: channel buffer size must not be negative")
	}

	if c.IsNone() {
		return nil
	}

	switch c.Kind {
	case AdapterKindFilesystem:
		if c.Filesystem == nil {
			return fmt.Errorf("persistence: filesystem config is required")
		}
		if c.Filesystem.Directory == "" {
			return fmt.Errorf("persistence: filesystem directory is required")
		}
	case AdapterKindPostgreSQL:
		if c.PostgreSQL == nil {
			return fmt.Errorf("persistence: postgresql config is required")
		}
		if c.PostgreSQL.DSN != "" {
			return nil
		}
		if c.PostgreSQL.Host == "" {
			return fmt.Errorf("persistence: postgresql host is required when DSN is not set")
		}
		if c.PostgreSQL.Port <= 0 {
			return fmt.Errorf("persistence: postgresql port must be positive when DSN is not set")
		}
		if c.PostgreSQL.User == "" {
			return fmt.Errorf("persistence: postgresql user is required when DSN is not set")
		}
		if c.PostgreSQL.Database == "" {
			return fmt.Errorf("persistence: postgresql database is required when DSN is not set")
		}
	default:
		// Allow registered custom adapters to validate their own configuration.
	}

	return nil
}
