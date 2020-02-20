package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalf("Usage: %s commit..commit [packages...]\n", os.Args[0])
	}
	commitRange := args[0]
	filterPackages := args[1:]

	localPackagesToDeps := packagesToDeps(filterPackages)
	changedLocalPackages := changedLocalPackages(commitRange)
	changedModules := changedModules(commitRange)

	for _, affectedPackage := range calcAffectedPackages(localPackagesToDeps, changedLocalPackages, changedModules) {
		fmt.Println(affectedPackage) // nolint:no-printf // inapplicable
	}
}

// Determines which of the given local packages (from the keys of localPackagesToDeps) were affected by changes to the given list of changedLocalPackages or changedModules
func calcAffectedPackages(localPackagesToDeps packagesToDepMap, changedLocalPackages, changedModules []string) []string {
	affectedPackageMap := make(map[string]bool)
	for _, pkg := range changedLocalPackages {
		// Any directly changed package is affected (obviously)
		affectedPackageMap[pkg] = true

		// Add all packages which depend on this changed package (including indirectly)
		for pkgPath, pkgDeps := range localPackagesToDeps {
			if pkgDeps[pkg] {
				affectedPackageMap[pkgPath] = true
			}
		}
	}

	for _, module := range changedModules {
		for pkgPath, pkgDeps := range localPackagesToDeps {
			for pkgDep := range pkgDeps {
				if strings.HasPrefix(pkgDep, module+"/") || pkgDep == module {
					// This local package was affected by a changed module or a subpackage of that module
					affectedPackageMap[pkgPath] = true
				}
			}
		}
	}

	affectedPackages := make([]string, 0, len(affectedPackageMap))
	for affectedPackage := range affectedPackageMap {
		affectedPackages = append(affectedPackages, affectedPackage)
	}
	sort.Strings(affectedPackages)
	return affectedPackages
}

func currentModule() string {
	cmd := exec.Command("go", "list", "-m")
	cmdOut, err := cmd.Output()
	if err != nil {
		log.Fatalf("Could not run git go list -m: %v", err)
	}

	return strings.TrimSpace(string(cmdOut))
}

// Returns a list of local packages which were changed in the given commit range.
// This normalizes all changed paths to be their Go import path, so:
// "service" becomes "<module root>/service"
func changedLocalPackages(commitRange string) []string {
	cmd := exec.Command("git", "diff", "--name-only", commitRange)
	cmdOut, err := cmd.Output()
	if err != nil {
		log.Fatalf("Could not run git diff: %v", err)
	}

	return calcChangedLocalPackages(string(cmdOut), currentModule())
}

func calcChangedLocalPackages(gitDiffOut, thisModule string) []string {
	packagesMap := make(map[string]bool)
	for _, f := range strings.Split(gitDiffOut, "\n") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if !strings.HasSuffix(f, ".go") {
			// skip non-Go files
			continue
		}
		firstDir := strings.Split(f, string(filepath.Separator))[0]
		if firstDir == "vendor" {
			// skip vendored files
			continue
		}

		pkg := filepath.Join(thisModule, filepath.Dir(f))
		packagesMap[pkg] = true
	}

	packages := make([]string, 0, len(packagesMap))
	for pkg := range packagesMap {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	return packages
}

func changedModules(commitRange string) []string {
	if _, err := os.Stat("go.sum"); os.IsNotExist(err) {
		// No go module here.
		return nil
	}

	cmd := exec.Command("git", "diff", commitRange, "go.sum")
	cmdOut, err := cmd.Output()
	if err != nil {
		log.Fatalf("Could not run git diff on go.sum: %v", err)
	}

	return calcChangedModules(string(cmdOut))
}

func calcChangedModules(gitDiffOut string) []string {
	moduleMap := make(map[string]bool)
	passedToFileHeader := false
	for _, line := range strings.Split(gitDiffOut, "\n") {
		// Skip lines until we're safely past the "+++ b/go.mod" line
		if !passedToFileHeader {
			if strings.HasPrefix(line, "+++") {
				passedToFileHeader = true
			}
			continue
		}

		if strings.HasPrefix(line, "+") {
			lineContents := strings.TrimSpace(line[1:])
			lineParts := strings.Split(lineContents, " ")
			if len(lineParts) == 3 {
				moduleMap[lineParts[0]] = true
			}
		}
	}

	modules := make([]string, 0, len(moduleMap))
	for mod := range moduleMap {
		modules = append(modules, mod)
	}
	sort.Strings(modules)

	return modules
}

type packagesToDepMap map[string]map[string]bool

// Gets a map of all local packages (anything not inside vendor) to a map of their dependencies, for easy lookups.
func packagesToDeps(filterPackages []string) packagesToDepMap {
	args := []string{"list", "-f", `{{.ImportPath}}|{{join .Deps ":"}}`}
	// Default to ./..., but if filters were given, limit our dependency graph to packages matching them.
	if len(filterPackages) == 0 {
		args = append(args, "./...")
	} else {
		args = append(args, filterPackages...)
	}
	cmd := exec.Command("go", args...)
	cmdOut, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error running go list: %s", err)
	}
	return calcPackagesToDeps(string(cmdOut))
}

func calcPackagesToDeps(goListOut string) packagesToDepMap {
	var result = make(packagesToDepMap)
	for _, pkgLine := range strings.Split(goListOut, "\n") {
		stringParts := strings.SplitN(pkgLine, "|", 2)
		importPath := stringParts[0]

		dependencyMap := make(map[string]bool)
		if len(stringParts) == 2 {
			dependencyPaths := strings.Split(stringParts[1], ":")
			for _, depPath := range dependencyPaths {
				dependencyMap[depPath] = true
			}
			result[importPath] = dependencyMap
		}
	}
	return result
}
