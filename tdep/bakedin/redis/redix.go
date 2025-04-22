package tdep_redis

import (
	"context"
	"fmt"
	"slices"

	"github.com/heffcodex/redix"

	"github.com/heffcodex/the/tdep"
)

func NewRedix[C redix.UniversalClient](config redix.Config, options ...tdep.Option) *tdep.D[C] {
	resolve := func(o tdep.OptSet) (C, error) {
		_config := redix.Config{
			Name:      config.Name,
			Namespace: config.Namespace,
			DSN:       config.DSN,
			Cert: redix.ConfigCert{
				Env:  config.Cert.Env,
				File: config.Cert.File,
				Data: slices.Clone(config.Cert.Data),
			},
		}

		if _config.Name == "" {
			_config.Name = o.Name()
		}

		if _config.Namespace == "" {
			_config.Namespace = redix.Namespace(o.Name()).Append(o.Env().String())
		}

		client, err := redix.NewClient(_config)
		if err != nil {
			return *new(C), err
		}

		return any(client).(C), nil //nolint:errcheck,revive // should never panic
	}

	return tdep.New(resolve, options...).WithHealthCheck(func(ctx context.Context, d *tdep.D[C]) error {
		instance, err := d.Get()
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}

		if !d.Options().IsSingleton() {
			defer func() { _ = d.Close(ctx) }()
		}

		if err = instance.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("ping: %w", err)
		}

		return nil
	})
}
