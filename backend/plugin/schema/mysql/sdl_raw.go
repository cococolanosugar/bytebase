package mysql

import (
	"cmp"
	"os"
	"slices"
	"strings"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
)

// useRawSDL reports whether the USE_RAW_DIFF environment variable is set to
// "true". When enabled, the SDL path uses SHOW CREATE TABLE/VIEW/TRIGGER
// output directly instead of re-synthesizing DDL from structured metadata.
func useRawSDL() bool {
	return strings.EqualFold(os.Getenv("USE_RAW_DIFF"), "true")
}

// tryBuildRawSDL attempts to assemble SDL text from SHOW CREATE output when
// USE_RAW_DIFF=true AND raw DDL data is available. Returns empty string
// otherwise, signaling the caller to fall back to the dumper.
func tryBuildRawSDL(metadata *storepb.DatabaseSchemaMetadata) string {
	if !useRawSDL() {
		return ""
	}
	return buildRawSDL(metadata)
}

// buildRawSDL assembles SDL text from SHOW CREATE output stored on each object.
// It returns an empty string when no raw DDL is available, signaling the caller
// to fall back to the existing dumper.
func buildRawSDL(metadata *storepb.DatabaseSchemaMetadata) string {
	sm := metadata.Schemas[0]

	// Check if we have raw DDL available for any table.
	hasRaw := false
	for _, table := range sm.Tables {
		if table.ShowCreate != "" {
			hasRaw = true
			break
		}
	}
	if !hasRaw {
		return ""
	}

	var buf strings.Builder

	// Tables — sorted by name for deterministic output.
	tables := make([]*storepb.TableMetadata, 0, len(sm.Tables))
	for _, table := range sm.Tables {
		if table.SkipDump {
			continue
		}
		tables = append(tables, table)
	}
	slices.SortFunc(tables, func(a, b *storepb.TableMetadata) int { return cmp.Compare(a.Name, b.Name) })
	for _, table := range tables {
		if table.ShowCreate != "" {
			buf.WriteString(table.ShowCreate)
			buf.WriteString(";\n\n")
		}
	}

	// Views — sorted by name, use ShowCreate when available.
	views := make([]*storepb.ViewMetadata, 0, len(sm.Views))
	for _, view := range sm.Views {
		if view.SkipDump {
			continue
		}
		views = append(views, view)
	}
	slices.SortFunc(views, func(a, b *storepb.ViewMetadata) int { return cmp.Compare(a.Name, b.Name) })
	for _, view := range views {
		if view.ShowCreate != "" {
			buf.WriteString(view.ShowCreate)
			buf.WriteString(";\n\n")
		}
	}

	// Functions — definition already stores the full CREATE FUNCTION text.
	functions := make([]*storepb.FunctionMetadata, 0, len(sm.Functions))
	for _, fn := range sm.Functions {
		if fn.SkipDump {
			continue
		}
		functions = append(functions, fn)
	}
	slices.SortFunc(functions, func(a, b *storepb.FunctionMetadata) int { return cmp.Compare(a.Name, b.Name) })
	for _, fn := range functions {
		if fn.Definition != "" {
			buf.WriteString(fn.Definition)
			buf.WriteString(";\n\n")
		}
	}

	// Procedures — definition already stores the full CREATE PROCEDURE text.
	procedures := make([]*storepb.ProcedureMetadata, 0, len(sm.Procedures))
	for _, proc := range sm.Procedures {
		if proc.SkipDump {
			continue
		}
		procedures = append(procedures, proc)
	}
	slices.SortFunc(procedures, func(a, b *storepb.ProcedureMetadata) int { return cmp.Compare(a.Name, b.Name) })
	for _, proc := range procedures {
		if proc.Definition != "" {
			buf.WriteString(proc.Definition)
			buf.WriteString(";\n\n")
		}
	}

	// Events — definition already stores the full CREATE EVENT text.
	events := make([]*storepb.EventMetadata, 0, len(sm.Events))
	events = append(events, sm.Events...)
	slices.SortFunc(events, func(a, b *storepb.EventMetadata) int { return cmp.Compare(a.Name, b.Name) })
	for _, event := range events {
		if event.Definition != "" {
			buf.WriteString(event.Definition)
			buf.WriteString(";\n\n")
		}
	}

	// Triggers — sorted by (table, trigger) for deterministic output.
	for _, table := range tables {
		triggers := make([]*storepb.TriggerMetadata, 0, len(table.Triggers))
		for _, trigger := range table.Triggers {
			if trigger.SkipDump {
				continue
			}
			triggers = append(triggers, trigger)
		}
		slices.SortFunc(triggers, func(a, b *storepb.TriggerMetadata) int { return cmp.Compare(a.Name, b.Name) })
		for _, trigger := range triggers {
			if trigger.ShowCreate != "" {
				buf.WriteString(trigger.ShowCreate)
				buf.WriteString(";\n\n")
			}
		}
	}

	return buf.String()
}
