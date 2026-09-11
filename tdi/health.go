package tdi

import "maps"

type HealthCheckResult struct {
	results  map[string]error
	errCount uint
}

func (r *HealthCheckResult) AllOk() bool {
	return r.errCount == 0
}

func (r *HealthCheckResult) Results() map[string]error {
	return maps.Clone(r.results)
}

func (r *HealthCheckResult) Errors() map[string]error {
	m := make(map[string]error, r.errCount)

	for name, err := range r.results {
		if err != nil {
			m[name] = err
		}
	}

	return m
}
