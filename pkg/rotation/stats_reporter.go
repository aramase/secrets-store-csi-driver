/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rotation

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	providerKey                 = "provider"
	errorKey                    = "error_type"
	osTypeKey                   = "os_type"
	rotatedKey                  = "rotated"
	rotationReconcileTotal      instrument.Int64Counter
	rotationReconcileErrorTotal instrument.Int64Counter
	rotationReconcileDuration   instrument.Float64Histogram
	runtimeOS                   = runtime.GOOS
)

type reporter struct{}

type StatsReporter interface {
	reportRotationCtMetric(ctx context.Context, provider string, wasRotated bool)
	reportRotationErrorCtMetric(ctx context.Context, provider, errType string, wasRotated bool)
	reportRotationDuration(ctx context.Context, duration float64)
}

func newStatsReporter() (StatsReporter, error) {
	var err error
	meter := global.Meter("secretsstore")
	if rotationReconcileTotal, err = meter.Int64Counter("total_rotation_reconcile", instrument.WithDescription("Total number of rotation reconciles")); err != nil {
		return nil, err
	}
	if rotationReconcileErrorTotal, err = meter.Int64Counter("total_rotation_reconcile_error", instrument.WithDescription("Total number of rotation reconciles with error")); err != nil {
		return nil, err
	}
	if rotationReconcileDuration, err = meter.Float64Histogram("rotation_reconcile_duration_sec", instrument.WithDescription("Distribution of how long it took to rotate secrets-store content for pods")); err != nil {
		return nil, err
	}
	return &reporter{}, nil
}

func (r *reporter) reportRotationCtMetric(ctx context.Context, provider string, wasRotated bool) {
	rotationReconcileTotal.Add(ctx, 1, attribute.String(providerKey, provider), attribute.String(osTypeKey, runtimeOS), attribute.Bool(rotatedKey, wasRotated))
}

func (r *reporter) reportRotationErrorCtMetric(ctx context.Context, provider, errType string, wasRotated bool) {
	rotationReconcileErrorTotal.Add(ctx, 1, attribute.String(providerKey, provider), attribute.String(errorKey, errType), attribute.String(osTypeKey, runtimeOS), attribute.Bool(rotatedKey, wasRotated))
}

func (r *reporter) reportRotationDuration(ctx context.Context, duration float64) {
	rotationReconcileDuration.Record(ctx, duration, attribute.String(osTypeKey, runtimeOS))
}
