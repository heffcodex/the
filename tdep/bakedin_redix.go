package tdep

import (
	"context"
	"fmt"
	"slices"

	"github.com/heffcodex/redix"
)

func NewRedix(config redix.Config, options ...Option) *D[*redix.Client] {
	resolve := func(o OptSet) (*redix.Client, error) {
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

		return redix.NewClient(_config)
	}

	return New(resolve, options...).WithHealthCheck(func(ctx context.Context, d *D[*redix.Client]) error {
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
