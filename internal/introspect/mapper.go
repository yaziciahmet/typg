package introspect

import (
	"fmt"
	"strings"

	"github.com/yaziciahmet/typg/internal/schema"
)

// ResolveType maps a postgres udt_name to a Go type.
func ResolveType(udtName string, enums map[string]schema.Enum) (schema.GoType, error) {
	if strings.HasPrefix(udtName, "_") {
		elem, err := ResolveType(udtName[1:], enums)
		if err != nil {
			return schema.GoType{}, err
		}
		return schema.GoType{
			Name:       "[]" + elem.Name,
			ImportPath: elem.ImportPath,
		}, nil
	}
	if _, ok := enums[udtName]; ok {
		return schema.GoType{Name: schema.ExportIdent(udtName), IsEnum: true}, nil
	}
	return mapScalarType(udtName)
}

func mapScalarType(udtName string) (schema.GoType, error) {
	switch strings.ToLower(udtName) {
	case "int2", "smallserial":
		return schema.GoType{Name: "int16"}, nil
	case "int4", "serial", "int":
		return schema.GoType{Name: "int32"}, nil
	case "int8", "bigserial":
		return schema.GoType{Name: "int64"}, nil
	case "bool":
		return schema.GoType{Name: "bool"}, nil
	case "text", "varchar", "bpchar", "name":
		return schema.GoType{Name: "string"}, nil
	case "timestamptz", "timestamp", "date":
		return schema.GoType{Name: "time.Time", ImportPath: "time"}, nil
	case "uuid":
		return schema.GoType{Name: "uuid.UUID", ImportPath: "uuid"}, nil
	case "json", "jsonb":
		return schema.GoType{Name: "json.RawMessage", ImportPath: "encoding/json"}, nil
	case "numeric", "decimal":
		return schema.GoType{Name: "decimal.Decimal", ImportPath: "github.com/shopspring/decimal"}, nil
	case "bytea":
		return schema.GoType{Name: "[]byte"}, nil
	case "float4", "float8":
		return schema.GoType{Name: "float64"}, nil
	default:
		return schema.GoType{}, fmt.Errorf("unsupported postgres type %q", udtName)
	}
}
