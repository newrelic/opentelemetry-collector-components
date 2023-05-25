package nrattributes

import (
	"context"
	"regexp"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

var (
	_ confmap.Converter = (*NRAttributesConverter)(nil)
)

type NRAttributesConverter struct {
	info component.BuildInfo
}

func New(info component.BuildInfo) confmap.Converter {
	return NRAttributesConverter{info}
}

func (bc NRAttributesConverter) Convert(ctx context.Context, conf *confmap.Conf) error {

	connectService := NewIdentityLookupService("abc")
	val := connectService.Lookup()

	println(val.GUID)

	// TODO check if resource processor is available ??

	const serviceExpr = "service(.+)processors"
	serviceEntryRe := regexp.MustCompile(serviceExpr)

	out := map[string]interface{}{
		"processors": map[string]interface{}{
			"resource/nr": map[string]interface{}{
				"attributes": []map[string]string{
					{
						"key":    "collector.name",
						"value":  bc.info.Command,
						"action": "insert",
					},
				},
			},
		},
	}

	for _, key := range conf.AllKeys() {
		if serviceEntryRe.MatchString(key) {
			out[key] = addNRattributes(conf.Get(key))
		}
	}

	return conf.Marshal(out)
}

func addNRattributes(value any) any {
	switch v := value.(type) {
	case []any:
		return append(v, "resource/nr")
	default:
		return v

	}
}
