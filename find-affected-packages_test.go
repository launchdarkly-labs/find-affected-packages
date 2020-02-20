package main

import (
	"reflect"
	"testing"
)

func TestCalcAffectedPackages(t *testing.T) {
	packagesToDeps := packagesToDepMap{
		"github.com/orgA/repoA/subPkg1": map[string]bool{
			"fmt":                   true,
			"github.com/orgB/repoB": true,
		},
		"github.com/orgA/repoA/subPkg2": map[string]bool{
			"fmt":                          true,
			"github.com/orgB/repoB/subPkg": true,
		},
		"github.com/orgA/repoA/subPkg3": map[string]bool{
			"fmt": true,
		},
		"github.com/orgA/repoA/subPkg4": map[string]bool{
			"github.com/orgC/repoC": true,
		},
		"github.com/orgA/repoA/subPkg5": map[string]bool{
			"fmt":                           true,
			"github.com/orgA/repoA/subPkg3": true,
		},
	}
	changedLocalPackages := []string{"github.com/orgA/repoA/subPkg3"}
	changedModules := []string{"github.com/orgB/repoB"}

	affectedPackages := calcAffectedPackages(packagesToDeps, changedLocalPackages, changedModules)
	expected := []string{
		"github.com/orgA/repoA/subPkg1", // Because it depends on a changed module's top-level package
		"github.com/orgA/repoA/subPkg2", // Because it depends on a subpackage of a changed module
		"github.com/orgA/repoA/subPkg3", // Because it was changed directly
		"github.com/orgA/repoA/subPkg5", // Because it depends on a subPkg3 which was changed directly
	}
	if !reflect.DeepEqual(affectedPackages, expected) {
		t.Logf("affectedPackages was %v, expected %v", affectedPackages, expected)
		t.Fail()
	}
}

func TestCalcChangedLocalPackages(t *testing.T) {
	diffOutput := `
.circleci/config.yml
Makefile
docs/readme.md
go.mod
go.sum
vendor/github.com/hashicorp/hcl/parse.go
vendor-machine/main.go
widget/config.go
`
	changedPackages := calcChangedLocalPackages(diffOutput, "test.com/this-module")
	expected := []string{
		"test.com/this-module/vendor-machine",
		"test.com/this-module/widget",
	}
	if !reflect.DeepEqual(changedPackages, expected) {
		t.Logf("changedPackages was %v, expected %v", changedPackages, expected)
		t.Fail()
	}
}

func TestCalcChangedModules(t *testing.T) {
	diffOutput := `
diff --git a/go.sum b/go.sum
index aaabbbccc..dddeeefff 100644
--- a/go.sum
+++ b/go.sum
@@ -1,4 +1,5 @@
 github.com/orgA/repoA v0.26.0/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
+github.com/orgA/repoB v0.3.1 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoC v0.3.1/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoD v0.0.0-20160522181843-27f122750802/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoE v2.2.0+incompatible/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
@@ -141,6 +142,7 @@ github.com/golang/mock v1.1.1 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoE v1.1.1/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoF v1.2.0 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoF v1.2.0/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
+github.com/orgB/repoF v1.3.1 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoF v1.3.1/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoG v0.0.0-20180518054509-2e65f85255db h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoG v0.0.0-20180518054509-2e65f85255db/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
@@ -148,6 +150,7 @@ github.com/gomodule/redigo v0.0.0-20190226174433-b47395aa1766 h1:
 github.com/orgB/repoG v0.0.0-20190226174433-b47395aa1766/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoH v0.0.0-20180813153112-4030bb1f1f0c h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoH v0.0.0-20180813153112-4030bb1f1f0c/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
+github.com/orgB/repoH v1.0.0 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
 github.com/orgB/repoH v1.0.0/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
@@ -265,8 +268,8 @@ github.com/magiconair/properties v1.8.1 h1:
 github.com/orgB/repoI v0.7.0/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
-github.com/orgC/repoJ v0.0.0-20150715184805-fe6ea2c8e398 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
-github.com/orgC/repoJ v0.0.0-20150715184805-fe6ea2c8e398/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
+github.com/orgC/repoJ v0.0.0-20150518220244-c574f6c50c85 h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
+github.com/orgC/repoJ v0.0.0-20150518220244-c574f6c50c85/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
-github.com/orgC/repoK v0.0.9/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
-github.com/orgC/repoL v0.0.3/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
-github.com/orgC/repoM v1.0.1/go.mod h1:WXkYYl6Yr3qBf1K79EBnL4mak0OimBfB0XUf9Vl28OQ=
`
	changedModules := calcChangedModules(diffOutput)
	expected := []string{
		"github.com/orgA/repoB",
		"github.com/orgB/repoF",
		"github.com/orgB/repoH",
		"github.com/orgC/repoJ",
	}
	if !reflect.DeepEqual(changedModules, expected) {
		t.Logf("changedModules was %v, expected %v", changedModules, expected)
		t.Fail()
	}
}

func TestCalcPackagesToDeps(t *testing.T) {
	goListOutput := `
github.com/orgA/repoA|strings:github.com/orgA/repoB
github.com/orgA/repoA/subPkg|fmt
`
	result := calcPackagesToDeps(goListOutput)
	expected := packagesToDepMap{
		"github.com/orgA/repoA": map[string]bool{
			"strings":               true,
			"github.com/orgA/repoB": true,
		},
		"github.com/orgA/repoA/subPkg": map[string]bool{
			"fmt": true,
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Logf("result was %v, expected %v", result, expected)
		t.Fail()
	}
}
