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
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	providerKey = "provider"
	errorKey    = "error_type"
	osTypeKey   = "os_type"
	rotatedKey  = "rotated"
	runtimeOS   = runtime.GOOS
)

type reporter struct {
	rotationReconcileTotal      instrument.Int64Counter
	rotationReconcileErrorTotal instrument.Int64Counter
	rotationReconcileDuration   instrument.Float64Histogram
}

type StatsReporter interface {
	reportRotationCtMetric(ctx context.Context, provider string, wasRotated bool)
	reportRotationErrorCtMetric(ctx context.Context, provider, errType string, wasRotated bool)
	reportRotationDuration(ctx context.Context, duration float64)
}

func newStatsReporter() (StatsReporter, error) {
	var err error
	r := &reporter{}
	meter := global.Meter("rotation")

	if r.rotationReconcileTotal, err = meter.Int64Counter("total_rotation_reconcile", instrument.WithDescription("Total number of rotation reconciles")); err != nil {
		return nil, err
	}
	if r.rotationReconcileErrorTotal, err = meter.Int64Counter("total_rotation_reconcile_error", instrument.WithDescription("Total number of rotation reconciles with error")); err != nil {
		return nil, err
	}
	if r.rotationReconcileDuration, err = meter.Float64Histogram("rotation_reconcile_duration_sec", instrument.WithDescription("Distribution of how long it took to rotate secrets-store content for pods")); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *reporter) reportRotationCtMetric(ctx context.Context, provider string, wasRotated bool) {
	opt := api.WithAttributes(
		attribute.Key(providerKey).String(provider),
		attribute.Key(osTypeKey).String(runtimeOS),
		attribute.Key(rotatedKey).Bool(wasRotated),
	)
	r.rotationReconcileTotal.Add(ctx, 1, opt)
}

func (r *reporter) reportRotationErrorCtMetric(ctx context.Context, provider, errType string, wasRotated bool) {
	opt := api.WithAttributes(
		attribute.Key(providerKey).String(provider),
		attribute.Key(errorKey).String(errType),
		attribute.Key(osTypeKey).String(runtimeOS),
		attribute.Key(rotatedKey).Bool(wasRotated),
	)
	r.rotationReconcileErrorTotal.Add(ctx, 1, opt)
}

func (r *reporter) reportRotationDuration(ctx context.Context, duration float64) {
	opt := api.WithAttributes(
		attribute.Key(osTypeKey).String(runtimeOS),
	)
	r.rotationReconcileDuration.Record(ctx, duration, opt)
}
