package introspect

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yaziciahmet/typg/internal/schema"
)

// Options configures postgres introspection.
type Options struct {
	Schema        string
	ExcludeTables []string
}

// Postgres reads schema metadata from a live database.
func Postgres(ctx context.Context, pool *pgxpool.Pool, opts Options) (schema.Database, error) {
	schemaName := opts.Schema
	if schemaName == "" {
		schemaName = "public"
	}
	excluded := excludeSet(opts.ExcludeTables)

	enums, err := loadEnums(ctx, pool, schemaName)
	if err != nil {
		return schema.Database{}, err
	}
	enumMap := enumIndex(enums)

	tables, err := loadTables(ctx, pool, schemaName, excluded, enumMap)
	if err != nil {
		return schema.Database{}, err
	}

	return schema.Database{Schema: schemaName, Enums: enums, Tables: tables}, nil
}

func loadEnums(ctx context.Context, pool *pgxpool.Pool, schemaName string) ([]schema.Enum, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.typname, e.enumlabel
		FROM pg_type t
		JOIN pg_enum e ON e.enumtypid = t.oid
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE n.nspname = $1
		ORDER BY t.typname, e.enumsortorder
	`, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byName := map[string][]string{}
	var order []string
	for rows.Next() {
		var name, label string
		if err := rows.Scan(&name, &label); err != nil {
			return nil, err
		}
		if _, ok := byName[name]; !ok {
			order = append(order, name)
		}
		byName[name] = append(byName[name], label)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	enums := make([]schema.Enum, 0, len(order))
	for _, name := range order {
		enums = append(enums, schema.Enum{Name: name, Values: byName[name]})
	}
	return enums, nil
}

func enumIndex(enums []schema.Enum) map[string]schema.Enum {
	m := make(map[string]schema.Enum, len(enums))
	for _, e := range enums {
		m[e.Name] = e
	}
	return m
}

func loadTables(ctx context.Context, pool *pgxpool.Pool, schemaName string, excluded map[string]bool, enums map[string]schema.Enum) ([]schema.Table, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.table_name, c.column_name, c.udt_name, c.is_nullable
		FROM information_schema.columns c
		JOIN information_schema.tables t
		  ON t.table_schema = c.table_schema AND t.table_name = c.table_name AND t.table_type = 'BASE TABLE'
		WHERE c.table_schema = $1
		ORDER BY c.table_name, c.ordinal_position
	`, schemaName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []schema.Table
	for rows.Next() {
		var table, name, udt, nullable string
		if err := rows.Scan(&table, &name, &udt, &nullable); err != nil {
			return nil, err
		}
		if excluded[table] {
			continue
		}
		goType, err := ResolveType(udt, enums)
		if err != nil {
			return nil, fmt.Errorf("table %s column %s: %w", table, name, err)
		}
		if len(tables) == 0 || tables[len(tables)-1].Name != table {
			tables = append(tables, schema.Table{Name: table})
		}
		last := &tables[len(tables)-1]
		last.Columns = append(last.Columns, schema.Column{
			Name:     name,
			PGType:   udt,
			Nullable: nullable == "YES",
			Go:       goType,
		})
	}
	return tables, rows.Err()
}

func excludeSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			out[n] = true
		}
	}
	return out
}
