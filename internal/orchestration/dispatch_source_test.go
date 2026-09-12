package orchestration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// This offline source guard prevents result collection from executing an action
// again after doRun has already dispatched it. It is not an AWS integration test.
func TestRunDispatchesThroughDoRunOnly(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "orchestrator.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var run *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == "Run" {
			run = function
			break
		}
	}
	if run == nil {
		t.Fatal("Run method not found")
	}
	dispatchCalls := 0
	ast.Inspect(run.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != "orch" {
			return true
		}
		switch selector.Sel.Name {
		case "doRun":
			dispatchCalls++
		case "Provision", "Unprovision", "Suspend":
			t.Errorf("Run must not invoke %s directly after dispatch", selector.Sel.Name)
		}
		return true
	})
	if dispatchCalls != 1 {
		t.Fatalf("expected one doRun dispatch site, got %d", dispatchCalls)
	}
}
