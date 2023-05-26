package entityguidprocessor

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

const (
	key = "entity.guid"
)

type entityGuidProcessor struct {
	guid string
	uri  string
}

func (eg *entityGuidProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	var err error
	if eg.guid == "" {
		eg.guid, err = getEntityGuid(eg.uri)
		if err != nil {
			return md, err
		}
	}
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		attrs := rms.At(i).Resource().Attributes()
		if _, found := attrs.Get(key); found {
			continue
		}
		attrs.PutStr(key, eg.guid)
	}
	return md, nil
}
