// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package proto // import "go.opentelemetry.io/collector/internal/cmd/pdatagen/internal/proto"

import (
	"fmt"

	"go.opentelemetry.io/collector/internal/cmd/pdatagen/internal/tmplutil"
)

// Validation templates generate the eager pass that UnmarshalProto performs
// before it stores the caller's byte slice for lazy decoding. The generated
// code consumes every field exactly as decodeProto would, including recursively
// validating embedded messages, but it discards the decoded values. That keeps
// future getter-triggered decoding panic-free without allocating message trees
// during UnmarshalProto.

const validateProtoFloat = `{{ if .repeated -}}
	case {{ .protoFieldID }}:
		switch wireType {
		case proto.WireTypeLen:
			var length int
			length, pos, err = proto.ConsumeLen(buf, pos)
			if err != nil {
				return err
			}
			if length%{{ div .bitSize 8 }} != 0 {
				return fmt.Errorf("proto: invalid field len = %d for field {{ .fieldName }}", length)
			}
		case proto.WireTypeI{{ .bitSize }}:
			_, pos, err = proto.ConsumeI{{ .bitSize }}(buf, pos)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("proto: wrong wireType = %d for field {{ .fieldName }}", wireType)
		}
{{- else }}
	case {{ .protoFieldID }}:
		if wireType != proto.WireTypeI{{ .bitSize }} {
			return fmt.Errorf("proto: wrong wireType = %d for field {{ .fieldName }}", wireType)
		}
		_, pos, err = proto.ConsumeI{{ .bitSize }}(buf, pos)
		if err != nil {
			return err
		}
{{- end }}`

const validateProtoFixed = validateProtoFloat

const validateProtoBool = `{{ if .repeated -}}
	case {{ .protoFieldID }}:
		switch wireType {
		case proto.WireTypeLen:
			var length int
			length, pos, err = proto.ConsumeLen(buf, pos)
			if err != nil {
				return err
			}
			startPos := pos - length
			for startPos < pos {
				_, startPos, err = proto.ConsumeVarint(buf[:pos], startPos)
				if err != nil {
					return err
				}
			}
		case proto.WireTypeVarint:
			_, pos, err = proto.ConsumeVarint(buf, pos)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("proto: wrong wireType = %d for field {{ .fieldName }}", wireType)
		}
{{- else }}
	case {{ .protoFieldID }}:
		if wireType != proto.WireTypeVarint {
			return fmt.Errorf("proto: wrong wireType = %d for field {{ .fieldName }}", wireType)
		}
		_, pos, err = proto.ConsumeVarint(buf, pos)
		if err != nil {
			return err
		}
{{- end }}`

const validateProtoVarint = validateProtoBool

const validateProtoBytesString = `
	case {{ .protoFieldID }}:
		if wireType != proto.WireTypeLen {
			return fmt.Errorf("proto: wrong wireType = %d for field {{ .fieldName }}", wireType)
		}
		_, pos, err = proto.ConsumeLen(buf, pos)
		if err != nil {
			return err
		}`

const validateProtoMessage = `
	case {{ .protoFieldID }}:
		if wireType != proto.WireTypeLen {
			return fmt.Errorf("proto: wrong wireType = %d for field {{ .fieldName }}", wireType)
		}
		var length int
		length, pos, err = proto.ConsumeLen(buf, pos)
		if err != nil {
			return err
		}
		startPos := pos - length
		err = validate{{ .messageName }}Proto(buf[startPos:pos])
		if err != nil {
			return err
		}`

func (pf *Field) GenValidateProto() string {
	tf := pf.getTemplateFields()
	switch pf.Type {
	case TypeDouble, TypeFloat:
		return tmplutil.Execute(tmplutil.Parse("validateProtoFloat", []byte(validateProtoFloat)), tf)
	case TypeFixed64, TypeSFixed64, TypeFixed32, TypeSFixed32:
		return tmplutil.Execute(tmplutil.Parse("validateProtoFixed", []byte(validateProtoFixed)), tf)
	case TypeInt32, TypeInt64, TypeUint32, TypeUint64, TypeEnum, TypeSInt32, TypeSInt64:
		return tmplutil.Execute(tmplutil.Parse("validateProtoVarint", []byte(validateProtoVarint)), tf)
	case TypeBool:
		return tmplutil.Execute(tmplutil.Parse("validateProtoBool", []byte(validateProtoBool)), tf)
	case TypeBytes, TypeString:
		return tmplutil.Execute(tmplutil.Parse("validateProtoBytesString", []byte(validateProtoBytesString)), tf)
	case TypeMessage:
		return tmplutil.Execute(tmplutil.Parse("validateProtoMessage", []byte(validateProtoMessage)), tf)
	}
	panic(fmt.Sprintf("unhandled case %T", pf.Type))
}
