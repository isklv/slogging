package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/isklv/slogging"
)

var TraceExemplar = func(ctx context.Context) prometheus.Labels {
	traceID, ok := ctx.Value(slogging.XB3TraceID).(string)
	if !ok || traceID == "" {
		return nil
	}

	return prometheus.Labels{
		slogging.XB3TraceID: traceID,
	}
}
