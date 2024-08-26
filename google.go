package otelutil

import (
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
)

func GoogleExporter(projectId string) (SpanExporter, error) {
	// Create exporter.
	exporter, err := texporter.New(texporter.WithProjectID(projectId))
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
