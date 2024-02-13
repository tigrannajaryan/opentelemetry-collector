// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/pdata/internal/cmd/pdatagen/internal"

var pentity = &Package{
	name: "pentity",
	path: "pentity",
	imports: []string{
		`"sort"`,
		``,
		`"go.opentelemetry.io/collector/pdata/internal"`,
		`"go.opentelemetry.io/collector/pdata/internal/data"`,
		`otlpentities "go.opentelemetry.io/collector/pdata/internal/data/protogen/entities/v1"`,
		`"go.opentelemetry.io/collector/pdata/pcommon"`,
	},
	testImports: []string{
		`"testing"`,
		`"unsafe"`,
		``,
		`"github.com/stretchr/testify/assert"`,
		``,
		`"go.opentelemetry.io/collector/pdata/internal"`,
		`"go.opentelemetry.io/collector/pdata/internal/data"`,
		`otlpentities "go.opentelemetry.io/collector/pdata/internal/data/protogen/entities/v1"`,
		`"go.opentelemetry.io/collector/pdata/pcommon"`,
	},
	structs: []baseStruct{
		scopeEntitiesSlice,
		scopeEntities,
		entitieslice,
		entityState,
	},
}

var scopeEntitiesSlice = &sliceOfPtrs{
	structName: "ScopeEntitiesSlice",
	element:    scopeEntities,
}

var scopeEntities = &messageValueStruct{
	structName:     "ScopeEntities",
	description:    "// ScopeEntities is a collection of entities from a LibraryInstrumentation.",
	originFullName: "otlpentities.ScopeEntities",
	fields: []baseField{
		scopeField,
		schemaURLField,
		&sliceField{
			fieldName:   "EntityStates",
			returnSlice: entitieslice,
		},
	},
}

var entitieslice = &sliceOfPtrs{
	structName: "EntityStateSlice",
	element:    entityState,
}

var entityState = &messageValueStruct{
	structName:     "EntityState",
	description:    "// EntityState are experimental implementation of OpenTelemetry Entity Data Model.\n",
	originFullName: "otlpentities.EntityState",
	fields: []baseField{
		&primitiveTypedField{
			fieldName:       "Timestamp",
			originFieldName: "TimeUnixNano",
			returnType:      timestampType,
		},
		&primitiveField{
			fieldName:  "Type",
			returnType: "string",
			defaultVal: `""`,
			testVal:    `"service"`,
		},
		entityId,
		attributes,
		droppedAttributesCount,
	},
}

var entityId = &sliceField{
	fieldName:   "Id",
	returnSlice: mapStruct,
}
