package tdep

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"
)

type BunConfig struct {
	DSN            string `mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	MaxConnections int    `mapstructure:"maxConnections" json:"maxConnections" yaml:"maxConnections"`
	MaxIdleTime    int    `mapstructure:"maxIdleTime" json:"maxIdleTime" yaml:"maxIdleTime"`
}

func (c *BunConfig) MaxIdleTimeSeconds() time.Duration {
	if c.MaxIdleTime < 1 {
		return 0
	}

	return time.Duration(c.MaxIdleTime) * time.Second
}

func NewBunPostgres(
	cfg BunConfig,
	onTuneConnector func(conn *pgdriver.Connector),
	onTuneSQLDB func(db *sql.DB),
	onTuneBunDB func(db *bun.DB),
	options ...Option,
) *D[*bun.DB] {
	resolve := func(o OptSet) (*bun.DB, error) {
		connOpts := []pgdriver.Option{
			pgdriver.WithApplicationName(o.Name()),
			pgdriver.WithDSN(cfg.DSN),
		}

		conn := pgdriver.NewConnector(connOpts...)
		if onTuneConnector != nil {
			onTuneConnector(conn)
		}

		sqlDB := sql.OpenDB(conn)
		sqlDB.SetMaxOpenConns(cfg.MaxConnections)
		sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTimeSeconds())

		if onTuneSQLDB != nil {
			onTuneSQLDB(sqlDB)
		}

		bunDB := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())
		if onTuneBunDB != nil {
			onTuneBunDB(bunDB)
		}

		logLevel := zap.ErrorLevel
		if o.IsDebug() {
			logLevel = zap.DebugLevel
		}

		stdLog, _ := zap.NewStdLogAt(o.Log(), logLevel)

		bunDB.AddQueryHook(
			bundebug.NewQueryHook(
				bundebug.WithVerbose(o.IsDebug()),
				bundebug.WithWriter(stdLog.Writer()),
			),
		)

		return bunDB, nil
	}

	return New(resolve, options...).WithHealthCheck(func(ctx context.Context, d *D[*bun.DB]) error {
		instance, err := d.Get()
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		if !d.Options().IsSingleton() {
			defer func() { _ = d.Close(ctx) }()
		}

		if err = instance.PingContext(ctx); err != nil {
			return fmt.Errorf("ping: %w", err)
		}

		return nil
	})
}
