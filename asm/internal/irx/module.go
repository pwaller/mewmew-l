package irx

import (
	"fmt"

	"github.com/mewmew/l/asm/ast"
	"github.com/mewmew/l/ir"
	"github.com/mewmew/l/ir/irutil"
	"github.com/mewmew/l/ir/metadata"
	"github.com/mewmew/l/ir/types"
)

// Translate translates the AST of the given module to an equivalent LLVM IR
// module.
func Translate(module *ast.Module) (*ir.Module, error) {
	// Create maps from identifier to definition.
	m := NewModule()
	m.indexIdents(module.Entities)

	// === [ Per module resolution ] ===

	// Resolve type definitions.
	resolveType := func(n interface{}) {
		if t, ok := n.(*types.Type); ok {
			if u, ok := (*t).(*types.NamedType); ok {
				if u.Type == nil {
					*t = m.typeDefs[u.Name]
				}
			}
		}
	}
	irutil.Walk(m.Module, resolveType)

	// Resolve comdat definitions.
	//
	//    *ast.ComdatName  -> *ir.ComdatDef

	// TODO: resolve comdats.

	// Resolve global variables, indirect symbols and functions.
	//
	//    *ast.GlobalIdent -> loop up in map. (*ir.Global, *ir.IndirectSymbol, *ir.Function)

	// TODO: resolve global variables.

	// Resolve attribute group definitions.
	//
	//    *ast.AttrGroupID -> *ir.AttrGroupDef

	// TODO: resolve attribute group definitions.

	// Resolve metadata definitions.
	//
	//    *ast.MetadataID  -> *ir.MetadataDef
	resolveMetadata := func(n interface{}) {
		switch n := n.(type) {
		case *metadata.MetadataNode:
			if i, ok := (*n).(*ast.MetadataID); ok {
				*n = m.metadataDefs[i.ID]
			}
		case *metadata.MDField:
			if i, ok := (*n).(*ast.MetadataID); ok {
				*n = m.metadataDefs[i.ID]
			}
		case *metadata.Metadata:
			if i, ok := (*n).(*ast.MetadataID); ok {
				*n = m.metadataDefs[i.ID]
			}
		case *metadata.MDNode:
			if i, ok := (*n).(*ast.MetadataID); ok {
				*n = m.metadataDefs[i.ID]
			}
		case *metadata.IntOrMDField:
			if i, ok := (*n).(*ast.MetadataID); ok {
				*n = m.metadataDefs[i.ID]
			}
		}
	}
	irutil.Walk(m.Module, resolveMetadata)

	// Resolve constants.
	//
	//    *ast.TypeConst

	// === [ Per function resolution ] ===

	// Resolve local variables (per function).
	//
	//    *ast.LocalIdent  -> look up in map. (*ir.BasicBlock, *ir.Param, *ir.LocalDef)
	//    *ast.TypeValue
	for _, f := range m.Funcs {
		_ = f
		// TODO: resolve local variables.
	}

	return m.Module, nil
}

// indexIdents indexes top-level entity definitions.
func (m *Module) indexIdents(entities []ast.TopLevelEntity) {
	// Create maps from identifier to definition.
	for _, entity := range entities {
		switch entity := entity.(type) {
		case *ir.SourceFilename:
			// Source filename.
			m.SourceFilename = entity
		case *ir.DataLayout:
			// Data layout.
			m.DataLayout = entity
		case *ir.TargetTriple:
			// Target triple.
			m.TargetTriple = entity
		case *ir.ModuleAsm:
			// Module-level inline assembly.
			m.ModuleAsms = append(m.ModuleAsms, entity)
		case *types.NamedType:
			// Type definitions.
			m.TypeDefs = append(m.TypeDefs, entity)
			ident := entity.Name
			if prev, ok := m.typeDefs[ident]; ok {
				panic(fmt.Errorf("type identifier %q already present; prev `%v`, new `%v`", ident, prev, entity))
			}
			m.typeDefs[ident] = entity
		case *ir.ComdatDef:
			// Comdat definitions.
			m.ComdatDefs = append(m.ComdatDefs, entity)
			name := entity.Name
			if prev, ok := m.comdatDefs[name]; ok {
				panic(fmt.Errorf("comdat name %q already present; prev `%v`, new `%v`", name, prev, entity))
			}
			m.comdatDefs[name] = entity
		case *ir.Global:
			// Global declarations and definitions.
			m.Globals = append(m.Globals, entity)
			ident := entity.Name
			if prev, ok := m.global(ident); ok {
				panic(fmt.Errorf("global identifier %q already present; prev `%v`, new `%v`", ident, prev, entity))
			}
			m.globals[ident] = entity
		case *ir.IndirectSymbol:
			// Indirect symbol definitions (aliases and IFuncs).
			m.IndirectSymbols = append(m.IndirectSymbols, entity)
			ident := entity.Name
			if prev, ok := m.global(ident); ok {
				panic(fmt.Errorf("global identifier %q already present; prev `%v`, new `%v`", ident, prev, entity))
			}
			m.indirectSymbols[ident] = entity
		case *ir.Function:
			// Function declarations and definitions.
			m.Funcs = append(m.Funcs, entity)
			ident := entity.Name
			if prev, ok := m.global(ident); ok {
				panic(fmt.Errorf("global identifier %q already present; prev `%v`, new `%v`", ident, prev, entity))
			}
			m.funcs[ident] = entity
		case *ir.AttrGroupDef:
			// Attribute group definitions.
			m.AttrGroupDefs = append(m.AttrGroupDefs, entity)
			id := entity.ID
			if prev, ok := m.attrGroupDefs[id]; ok {
				panic(fmt.Errorf("attribute group ID %q already present; prev `%v`, new `%v`", id, prev, entity))
			}
			m.attrGroupDefs[id] = entity
		case *metadata.NamedMetadataDef:
			// Named metadata definitions.
			m.NamedMetadataDefs = append(m.NamedMetadataDefs, entity)
			name := entity.Name
			if prev, ok := m.namedMetadataDefs[name]; ok {
				panic(fmt.Errorf("metadata name %q already present; prev `%v`, new `%v`", name, prev, entity))
			}
			m.namedMetadataDefs[name] = entity
		case *metadata.MetadataDef:
			// Metadata definitions.
			m.MetadataDefs = append(m.MetadataDefs, entity)
			id := entity.ID
			if prev, ok := m.metadataDefs[id]; ok {
				panic(fmt.Errorf("metadata ID %q already present; prev `%v`, new `%v`", id, prev, entity))
			}
			m.metadataDefs[id] = entity
		case *ir.UseListOrder:
			// Use-list order directives.
			m.UseListOrders = append(m.UseListOrders, entity)
		case *ir.UseListOrderBB:
			// Basic block specific use-list order directives.
			m.UseListOrderBBs = append(m.UseListOrderBBs, entity)
		default:
			panic(fmt.Errorf("support for top-level entity %T not yet implemented", entity))
		}
	}
}
