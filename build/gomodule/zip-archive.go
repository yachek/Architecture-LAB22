package gomodule

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

var (
	// Package context used to define Ninja build rules.
	//pctx = blueprint.NewPackageContext("github.com/vladShadow/Architecture-LAB22/build/gomodule")

	// Ninja rule to create zip archive
	zipArchive = pctx.StaticRule("zipArchive", blueprint.RuleParams{
		Command:     "cd $workDir && zip -r $outputPath $inputFiles",
		Description: "zip archive at $outputPath",
	}, "workDir", "outputPath", "inputFiles")
)

type zipArchiveModuleType struct {
	blueprint.SimpleName

	properties struct {
		Srcs        []string
		SrcsExclude []string
		Deps        []string
	}
}

func (gb *zipArchiveModuleType) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return gb.properties.Deps
}

func (gb *zipArchiveModuleType) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "archives", name)

	var inputs []string
	inputErors := false
	for _, src := range gb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, gb.properties.SrcsExclude); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	if inputErors {
		return
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Create %s zip archive", name),
		Rule:        zipArchive,
		Outputs:     []string{outputPath},
		//Implicits:   inputs,
		Args: map[string]string{
			"workDir":    ctx.ModuleDir(),
			"outputPath": outputPath,
			"inputFiles": strings.Join(inputs, " "),
		},
	})

}

// ZipArchiveFactory is a factory for zip archive module type which supports creating zip archive files
func ZipArchiveFactory() (blueprint.Module, []interface{}) {
	mType := &zipArchiveModuleType{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
