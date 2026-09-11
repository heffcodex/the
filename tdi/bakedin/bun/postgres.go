package tdi_bun

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"go.uber.org/zap"

	"github.com/heffcodex/the/tdi"
)

var _ IDB = (*bun.DB)(nil)

type IDB interface {
	bun.IDB
	PingContext(ctx context.Context) error
}

func NewPostgres[C IDB](
	cfg Config,
	onTuneConnector func(conn *pgdriver.Connector),
	onTuneSQLDB func(db *sql.DB),
	onTuneBunDB func(db *bun.DB),
	paramFuncs ...tdi.ParamFunc,
) *tdi.D[C] {
	resolve := func(_ context.Context, p tdi.Params) (C, error) {
		connOpts := []pgdriver.Option{
			pgdriver.WithApplicationName(p.Name()),
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
		if p.IsDebug() {
			logLevel = zap.DebugLevel
		}

		stdLog, _ := zap.NewStdLogAt(p.Log(), logLevel)

		bunDB.WithQueryHook(
			bundebug.NewQueryHook(
				bundebug.WithVerbose(p.IsDebug()),
				bundebug.WithWriter(stdLog.Writer()),
			),
		)

		return any(bunDB).(C), nil //nolint:errcheck,revive // should never panic
	}

	return tdi.NewWithHealthCheck(resolve, func(ctx context.Context, d *tdi.D[C]) error {
		instance, err := d.Get(ctx)
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		if err = instance.PingContext(ctx); err != nil {
			return fmt.Errorf("ping: %w", err)
		}

		return nil
	}, paramFuncs...)
}
