package otlptext

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func marshalResource(res pcommon.Resource, buf *dataBuffer) {
	buf.logAttributes("Resource attributes", res.Attributes())
	buf.logEntry("Entity Type: %s", res.EntityType())
	buf.logAttributes("Entity Id", res.EntityId())
}
