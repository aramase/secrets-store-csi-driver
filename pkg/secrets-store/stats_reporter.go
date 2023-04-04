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

package secretsstore

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var (
	providerKey             = "provider"
	errorKey                = "error_type"
	osTypeKey               = "os_type"
	nodePublishTotal        instrument.Int64Counter
	nodeUnPublishTotal      instrument.Int64Counter
	nodePublishErrorTotal   instrument.Int64Counter
	nodeUnPublishErrorTotal instrument.Int64Counter
	syncK8sSecretTotal      instrument.Int64Counter
	syncK8sSecretDuration   instrument.Float64Histogram
	runtimeOS               = runtime.GOOS
)

type reporter struct{}

type StatsReporter interface {
	ReportNodePublishCtMetric(ctx context.Context, provider string)
	ReportNodeUnPublishCtMetric(ctx context.Context)
	ReportNodePublishErrorCtMetric(ctx context.Context, provider, errType string)
	ReportNodeUnPublishErrorCtMetric(ctx context.Context)
	ReportSyncK8SecretCtMetric(ctx context.Context, provider string, count int)
	ReportSyncK8SecretDuration(ctx context.Context, duration float64)
}

func NewStatsReporter() (StatsReporter, error) {
	var err error
	meter := global.Meter("secretsstore")
	if nodePublishTotal, err = meter.Int64Counter("total_node_publish", instrument.WithDescription("Total number of node publish calls")); err != nil {
		return nil, err
	}
	if nodeUnPublishTotal, err = meter.Int64Counter("total_node_unpublish", instrument.WithDescription("Total number of node unpublish calls")); err != nil {
		return nil, err
	}
	if nodePublishErrorTotal, err = meter.Int64Counter("total_node_publish_error", instrument.WithDescription("Total number of node publish calls with error")); err != nil {
		return nil, err
	}
	if nodeUnPublishErrorTotal, err = meter.Int64Counter("total_node_unpublish_error", instrument.WithDescription("Total number of node unpublish calls with error")); err != nil {
		return nil, err
	}
	if syncK8sSecretTotal, err = meter.Int64Counter("total_sync_k8s_secret", instrument.WithDescription("Total number of k8s secrets synced")); err != nil {
		return nil, err
	}
	if syncK8sSecretDuration, err = meter.Float64Histogram("sync_k8s_secret_duration_sec", instrument.WithDescription("Distribution of how long it took to sync k8s secret")); err != nil {
		return nil, err
	}
	return &reporter{}, nil
}

func (r *reporter) ReportNodePublishCtMetric(ctx context.Context, provider string) {
	nodePublishTotal.Add(ctx, 1, attribute.String(providerKey, provider), attribute.String(osTypeKey, runtimeOS))
}

func (r *reporter) ReportNodeUnPublishCtMetric(ctx context.Context) {
	nodeUnPublishTotal.Add(ctx, 1, attribute.String(osTypeKey, runtimeOS))
}

func (r *reporter) ReportNodePublishErrorCtMetric(ctx context.Context, provider, errType string) {
	nodePublishErrorTotal.Add(ctx, 1, attribute.String(providerKey, provider), attribute.String(errorKey, errType), attribute.String(osTypeKey, runtimeOS))
}

func (r *reporter) ReportNodeUnPublishErrorCtMetric(ctx context.Context) {
	nodeUnPublishErrorTotal.Add(ctx, 1, attribute.String(osTypeKey, runtimeOS))
}

func (r *reporter) ReportSyncK8SecretCtMetric(ctx context.Context, provider string, count int) {
	syncK8sSecretTotal.Add(ctx, int64(count), attribute.String(providerKey, provider), attribute.String(osTypeKey, runtimeOS))
}

func (r *reporter) ReportSyncK8SecretDuration(ctx context.Context, duration float64) {
	syncK8sSecretDuration.Record(ctx, duration, attribute.String(osTypeKey, runtimeOS))
}
