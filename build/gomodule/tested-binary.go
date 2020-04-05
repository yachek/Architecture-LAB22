package gomodule

import (
	"fmt"
	"path"
	"regexp"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/vladShadow/Architecture-LAB22/build/gomodule")

	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd $workDir && go mod vendor",
		Description: "vendor dependencies of $name",
	}, "workDir", "name")

	// Ninja rule to execute tests.
	goTest = pctx.StaticRule("test", blueprint.RuleParams{
		Command:     "cd $workDir && go test -v $pkg > $outputPath",
		Description: "test go command $pkg",
	}, "workDir", "outputPath", "pkg")
)

type goTestedBinaryModuleType struct {
	blueprint.SimpleName

	properties struct {
		Pkg         string
		TestPkg     string
		Srcs        []string
		SrcsExclude []string
		VendorFirst bool
		Deps        []string
	}
}

func (gb *goTestedBinaryModuleType) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return gb.properties.Deps
}

func (gb *goTestedBinaryModuleType) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "bin", name)
	outputPathTest := path.Join(config.BaseOutputDir, "bin", "test.txt")

	var inputsTest []string
	inputErors := false
	for _, src := range gb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, gb.properties.SrcsExclude); err == nil {
			inputsTest = append(inputsTest, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	inputs := inputsTest
	if inputErors {
		return
	}
	for i := 0; i < len(inputsTest); i++ {
		if val, _ := regexp.Match(".*_test\\.go$", []byte(inputsTest[i])); val == false {
			inputs = append(inputs, inputsTest[i])
		}
	}

	if gb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{outputPath},
		Implicits:   inputs,
		Args: map[string]string{
			"outputPath": outputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        gb.properties.Pkg,
		},
	})

	if gb.properties.TestPkg != "" {
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Test %s and save results", name),
			Rule:        goTest,
			Outputs:     []string{outputPathTest},
			Implicits:   inputsTest,
			Args: map[string]string{
				"outputPath": outputPathTest,
				"workDir":    ctx.ModuleDir(),
				"pkg":        gb.properties.TestPkg,
			},
		})
	}
}

// TestedBinFactory is a factory for go binary module type which supports Go command packages with running tests
func TestedBinFactory() (blueprint.Module, []interface{}) {
	mType := &goTestedBinaryModuleType{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
