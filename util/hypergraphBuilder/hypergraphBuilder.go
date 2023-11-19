// This was added by MichaelMartinek

package hypergraphBuilder

import (
	"github.com/moorara/algo/unionfind"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/opcode"
	"strconv"
	"strings"
)

type HypergraphBuilder struct {
	hg           map[string]map[string]bool // This is a Map[string]Set{string} as Map[string]Map[string]bool
	nextVar      int
	colToVar     map[string]int
	colToVarUsed map[string]bool // This is used to determine weather a key of colToVar is used or not, since the zero value in colToVar is 0, which is also a value which is used
	vars         unionfind.UnionFind
}

// NewHypergraphBuilder : Create a new HypergraphBuilder
// Note that the UnionFind vars datastructure still has to be initialized after this call
func NewHypergraphBuilder(numberOfJoinConditions int) HypergraphBuilder {
	return HypergraphBuilder{
		hg:           make(map[string]map[string]bool),
		nextVar:      0,
		colToVar:     make(map[string]int),
		colToVarUsed: make(map[string]bool),
		vars:         unionfind.NewQuickUnion(numberOfJoinConditions),
	}
}

// BuildEdgeInit : Build/Initialize an edge in the hypergraph --> for a table
func (h *HypergraphBuilder) BuildEdgeInit(table string) {
	h.hg[table] = make(map[string]bool)
}

// BuildEdge : Build an edge in the hypergraph --> for a table.column
func (h *HypergraphBuilder) BuildEdge(table string, col string) {
	h.hg[table][col] = true

	attr := stringify(table, col)

	if h.colToVarUsed[attr] == false {
		h.colToVarUsed[attr] = true
		h.colToVar[attr] = h.nextVar
		h.nextVar++
	}
}

// BuildJoin : Add a join to the hypergraph
func (h *HypergraphBuilder) BuildJoin(eq *ast.BinaryOperationExpr) {
	// check eq is of type eq
	if eq.Op != opcode.EQ {
		return
	}

	// check tables exist in hg, columns exist in hg
	left, ok := h.hg[eq.L.(*ast.ColumnNameExpr).Name.Table.L]
	right, ok2 := h.hg[eq.R.(*ast.ColumnNameExpr).Name.Table.L]
	if !ok || !ok2 {
		return
	}
	if left[eq.L.(*ast.ColumnNameExpr).Name.Name.L] == false || right[eq.R.(*ast.ColumnNameExpr).Name.Name.L] == false {
		return
	}

	// UnionFind Union call --> connect the two vertices
	h.vars.Union(
		h.colToVar[stringify(eq.L.(*ast.ColumnNameExpr).Name.Table.L, eq.L.(*ast.ColumnNameExpr).Name.Name.L)],
		h.colToVar[stringify(eq.R.(*ast.ColumnNameExpr).Name.Table.L, eq.R.(*ast.ColumnNameExpr).Name.Name.L)],
	)
}

// MakeHypergraph : Get the String representation of a hypergraph
func (h *HypergraphBuilder) MakeHypergraph() (ret []string) {
	for k, v := range h.hg {
		var sb strings.Builder
		sb.WriteString(k)
		sb.WriteString("(")

		l := len(v)
		ind := 1
		for vertex, _ := range v {

			vert := h.colToVar[stringify(k, vertex)]
			vertexInd, _ := h.vars.Find(vert)
			sb.WriteString("v" + strconv.Itoa(vertexInd))

			if ind < l {
				sb.WriteString(",")
			}
			ind++

		}

		sb.WriteString(")")
		ret = append(ret, sb.String())

	}
	return ret
}

// GetMapping : Get the Mapping of the hypergraph in string representation
func (h *HypergraphBuilder) GetMapping() (ret []string) {
	var varToCol = make(map[int][]string)

	for c, v := range h.colToVar {
		va, ok := h.vars.Find(v)
		if ok && len(varToCol[va]) == 0 {
			varToCol[va] = make([]string, 1)
			varToCol[va][0] = c
		} else if ok {
			varToCol[va] = append(varToCol[va], c)
		}
	}

	ret = make([]string, len(varToCol))
	retInd := 0

	for k, v := range varToCol {
		var sb strings.Builder
		sb.WriteString(strconv.Itoa(k))
		sb.WriteString("=")

		for i, e := range v {
			sb.WriteString(e)

			if i+1 < len(v) {
				sb.WriteString(",")
			}
		}
		ret[retInd] = sb.String()
		retInd++
	}
	return ret
}

// GetMappingAsMap : Get the Mapping of the hypergraph in map representation
func (h *HypergraphBuilder) GetMappingAsMap() map[int][]string {
	var varToCol = make(map[int][]string)

	for c, v := range h.colToVar {
		va, ok := h.vars.Find(v)
		if ok && len(varToCol[va]) == 0 {
			varToCol[va] = make([]string, 1)
			varToCol[va][0] = c
		} else if ok {
			varToCol[va] = append(varToCol[va], c)
		}
	}

	return varToCol
}

// stringify Create string table.col
func stringify(table string, col string) string {
	var sb strings.Builder

	sb.WriteString(table)
	sb.WriteString(".")
	sb.WriteString(col)

	return sb.String()
}
