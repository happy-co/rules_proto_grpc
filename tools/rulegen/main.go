// Wsifier, a tool to parse BUILD files and bzl files, generate tests cases and
// documentation.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/urfave/cli"
)

var defaultPlatforms = []string{"linux", "windows", "macos"}
var ciPlatforms = []string{"ubuntu1804", "windows", "macos"}
var ciPlatformsMap = map[string][]string{
	"linux":   []string{"ubuntu1604", "ubuntu1804", "rbe_ubuntu1604", "rbe_ubuntu1804"},
	"windows": []string{"windows"},
	"macos":   []string{"macos"},
}

func main() {
	app := cli.NewApp()
	app.Name = "rulegen"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "dir",
			Usage: "Directory to scan",
			Value: ".",
		},
		&cli.StringFlag{
			Name:  "header",
			Usage: "Template for the main readme header",
			Value: "tools/rulegen/README.header.md",
		},
		&cli.StringFlag{
			Name:  "footer",
			Usage: "Template for the main readme footer",
			Value: "tools/rulegen/README.footer.md",
		},
		&cli.StringFlag{
			Name:  "ref",
			Usage: "Version ref to use for main readme",
			Value: "{GIT_COMMIT_ID}",
		},
		&cli.StringFlag{
			Name:  "sha256",
			Usage: "Sha256 value to use for main readme",
			Value: "{ARCHIVE_TAR_GZ_SHA256}",
		},
		&cli.StringFlag{
			Name:  "github_url",
			Usage: "URL for github download",
			Value: "https://github.com/rules-proto-grpc/rules_proto_grpc/archive/{ref}.tar.gz",
		},
		&cli.StringFlag{
			Name:  "available_tests",
			Usage: "File containing the list of available routeguide tests",
			Value: "available_tests.txt",
		},
	}
	app.Action = func(c *cli.Context) error {
		err := action(c)
		if err != nil {
			return cli.NewExitError("%v", 1)
		}
		return nil
	}

	app.Run(os.Args)
}

func action(c *cli.Context) error {
	dir := c.String("dir")
	if dir == "" {
		return fmt.Errorf("--dir required")
	}

	ref := c.String("ref")
	sha256 := c.String("sha256")
	githubURL := c.String("github_url")

	// Autodetermine sha256 if we have a real commit and templated sha256 value
	if ref != "{GIT_COMMIT_ID}" && sha256 == "{ARCHIVE_TAR_GZ_SHA256}" {
		sha256 = mustGetSha256(strings.Replace(githubURL, "{ref}", ref, 1))
	}

	languages := []*Language{
		makeAndroid(),
		makeClosure(),
		makeCpp(),
		makeCsharp(),
		makeD(),
		makeGo(),
		makeJava(),
		makeNode(),
		makeObjc(),
		makePhp(),
		makePython(),
		makeRuby(),
		makeRust(),
		makeScala(),
		makeSwift(),

		makeGogo(),
		makeGrpcGateway(),
		makeGithubComGrpcGrpcWeb(),
	}

	for _, lang := range languages {
		mustWriteLanguageReadme(dir, lang)
		mustWriteLanguageDefs(dir, lang)
		mustWriteLanguageRules(dir, lang)
		mustWriteLanguageExamples(dir, lang)
	}

	mustWriteReadme(dir, c.String("header"), c.String("footer"), struct {
		Ref, Sha256 string
	}{
		Ref:    ref,
		Sha256: sha256,
	}, languages)

	mustWriteBazelciPresubmitYml(dir, languages, []string{}, c.String("available_tests"))

	mustWriteExamplesMakefile(dir, languages)
	mustWriteTestWorkspacesMakefile(dir)
	mustWriteHttpArchiveTestWorkspace(dir, ref, sha256)

	return nil
}

func mustWriteLanguageRules(dir string, lang *Language) {
	for _, rule := range lang.Rules {
		mustWriteLanguageRule(dir, lang, rule)
	}
}

func mustWriteLanguageRule(dir string, lang *Language, rule *Rule) {
	out := &LineWriter{}
	out.t(rule.Implementation, &ruleData{lang, rule})
	out.ln()
	out.MustWrite(filepath.Join(dir, lang.Dir, rule.Name+".bzl"))
}

func mustWriteLanguageExamples(dir string, lang *Language) {
	for _, rule := range lang.Rules {
		exampleDir := filepath.Join(dir, "example", lang.Dir, rule.Name)
		err := os.MkdirAll(exampleDir, os.ModePerm)
		if err != nil {
			log.Fatalf("FAILED to create %s: %v", exampleDir, err)
		}
		mustWriteLanguageExampleWorkspace(exampleDir, lang, rule)
		mustWriteLanguageExampleBuildFile(exampleDir, lang, rule)
		mustWriteLanguageExampleBazelrcFile(exampleDir, lang, rule)
	}
}

func mustWriteLanguageExampleWorkspace(dir string, lang *Language, rule *Rule) {
	out := &LineWriter{}
	depth := strings.Split(lang.Dir, "/")
	// +2 as we are in the example/{rule} subdirectory
	relpath := strings.Repeat("../", len(depth)+2)

	out.w(`local_repository(
    name = "rules_proto_grpc",
    path = "%s",
)

load("@rules_proto_grpc//:repositories.bzl", "rules_proto_grpc_toolchains", "rules_proto_grpc_repos")
rules_proto_grpc_toolchains()
rules_proto_grpc_repos()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")
rules_proto_dependencies()
rules_proto_toolchains()`, relpath)

	out.ln()
	out.t(rule.WorkspaceExample, &ruleData{lang, rule})
	out.ln()
	out.MustWrite(filepath.Join(dir, "WORKSPACE"))
}

func mustWriteLanguageExampleBuildFile(dir string, lang *Language, rule *Rule) {
	out := &LineWriter{}
	out.t(rule.BuildExample, &ruleData{lang, rule})
	out.ln()
	out.MustWrite(filepath.Join(dir, "BUILD.bazel"))
}

func mustWriteLanguageExampleBazelrcFile(dir string, lang *Language, rule *Rule) {
	out := &LineWriter{}
	for _, f := range lang.Flags {
		if f.Description != "" {
			out.w("# %s", f.Description)
		} else {
			out.w("#")
		}
		out.w("%s --%s=%s", f.Category, f.Name, f.Value)
	}
	for _, f := range rule.Flags {
		if f.Description != "" {
			out.w("# %s", f.Description)
		} else {
			out.w("#")
		}
		out.w("%s --%s=%s", f.Category, f.Name, f.Value)
	}
	out.ln()
	out.MustWrite(filepath.Join(dir, ".bazelrc"))
}

func mustWriteLanguageDefs(dir string, lang *Language) {
	out := &LineWriter{}
	out.w("# Aggregate all `%s` rules to one loadable file", lang.Name)
	for _, rule := range lang.Rules {
		out.w(`load(":%s.bzl", _%s="%s")`, rule.Name, rule.Name, rule.Name)
	}
	out.ln()
	for _, rule := range lang.Rules {
		out.w(`%s = _%s`, rule.Name, rule.Name)
	}
	out.ln()
	if len(lang.Aliases) > 0 {
		out.w(`# Aliases`)

		aliases := make([]string, 0, len(lang.Aliases))
		for alias := range lang.Aliases {
			aliases = append(aliases, alias)
		}
		sort.Strings(aliases)

		for _, alias := range aliases {
			out.w(`%s = _%s`, alias, lang.Aliases[alias])
		}

		out.ln()
	}
	out.MustWrite(filepath.Join(dir, lang.Dir, "defs.bzl"))
}

func mustWriteLanguageReadme(dir string, lang *Language) {
	out := &LineWriter{}

	out.w("# %s rules", lang.DisplayName)
	out.ln()

	if lang.Notes != nil {
		out.t(lang.Notes, lang)
		out.ln()
	}

	out.w("| Rule | Description |")
	out.w("| ---: | :--- |")
	for _, rule := range lang.Rules {
		out.w("| [%s](#%s) | %s |", rule.Name, rule.Name, rule.Doc)
	}
	out.ln()

	for _, rule := range lang.Rules {
		out.w(`---`)
		out.ln()
		out.w("## `%s`", rule.Name)
		out.ln()

		if rule.Experimental {
			out.w(`> NOTE: this rule is EXPERIMENTAL.  It may not work correctly or even compile!`)
			out.ln()
		}
		out.w(rule.Doc)
		out.ln()

		out.w("### `WORKSPACE`")
		out.ln()

		out.w("```starlark")
		out.t(rule.WorkspaceExample, &ruleData{lang, rule})
		out.w("```")
		out.ln()

		out.w("### `BUILD.bazel`")
		out.ln()

		out.w("```starlark")
		out.t(rule.BuildExample, &ruleData{lang, rule})
		out.w("```")
		out.ln()

		if len(rule.Flags) > 0 {
			out.w("### `Flags`")
			out.ln()

			out.w("| Category | Flag | Value | Description |")
			out.w("| --- | --- | --- | --- |")
			for _, f := range rule.Flags {
				out.w("| %s | %s | %s | %s |", f.Category, f.Name, f.Value, f.Description)
			}
			out.ln()
		}

		out.w("### Attributes")
		out.ln()
		out.w("| Name | Type | Mandatory | Default | Description |")
		out.w("| ---: | :--- | --------- | ------- | ----------- |")
		for _, attr := range rule.Attrs {
			out.w("| `%s` | `%s` | %t | `%s`    | %s          |", attr.Name, attr.Type, attr.Mandatory, attr.Default, attr.Doc)
		}
		out.ln()
	}

	out.MustWrite(filepath.Join(dir, lang.Dir, "README.md"))
}

func mustWriteReadme(dir, header, footer string, data interface{}, languages []*Language) {
	out := &LineWriter{}

	out.tpl(header, data)
	out.ln()

	out.w("## Rules")
	out.ln()

	out.w("| Language | Rule | Description")
	out.w("| ---: | :--- | :--- |")
	for _, lang := range languages {
		for _, rule := range lang.Rules {
			dirLink := fmt.Sprintf("[%s](/%s)", lang.DisplayName, lang.Dir)
			ruleLink := fmt.Sprintf("[%s](/%s#%s)", rule.Name, lang.Dir, rule.Name)
			exampleLink := fmt.Sprintf("[example](/example/%s/%s)", lang.Dir, rule.Name)
			out.w("| %s | %s | %s (%s) |", dirLink, ruleLink, rule.Doc, exampleLink)
		}
	}
	out.ln()

	out.tpl(footer, data)

	out.MustWrite(filepath.Join(dir, "README.md"))
}

func mustWriteBazelciPresubmitYml(dir string, languages []*Language, envVars []string, availableTestsPath string) {
	// Read available tests
	content, err := ioutil.ReadFile(availableTestsPath)
	if err != nil {
		log.Fatal(err)
	}
	availableTestLabels := strings.Split(string(content), "\n")

	// Write header
	out := &LineWriter{}
	out.w("---")
	out.w("tasks:")

	//
	// Write tasks for main code
	//
	for _, ciPlatform := range ciPlatforms {
		// Skip windows, due to issues with 'undeclared inclusion'
		if ciPlatform == "windows" {
			continue
		}
		out.w("  main_%s:", ciPlatform)
		out.w("    name: build & test all")
		out.w("    platform: %s", ciPlatform)
		out.w("    environment:")
		out.w(`      CC: clang`)
		if ciPlatform == "macos" {
			out.w("    build_flags:")
			out.w(`    - "--copt=-DGRPC_BAZEL_BUILD"`) // https://github.com/bazelbuild/bazel/issues/4341 required for macos
		}
		out.w("    build_targets:")
		for _, lang := range languages {
			// Skip experimental or excluded
			if doTestOnPlatform(lang, nil, ciPlatform) {
				out.w(`    - "//%s/..."`, lang.Dir)
			}
		}
		out.w("    test_flags:")
		if ciPlatform == "macos" {
			out.w(`    - "--copt=-DGRPC_BAZEL_BUILD"`) // https://github.com/bazelbuild/bazel/issues/4341 required for macos
		}
		out.w(`    - "--test_output=errors"`)
		out.w("    test_targets:")
		for _, clientLang := range languages {
			for _, serverLang := range languages {
				if doTestOnPlatform(clientLang, nil, ciPlatform) && doTestOnPlatform(serverLang, nil, ciPlatform) && stringInSlice(fmt.Sprintf("//example/routeguide:%s_%s", clientLang.Name, serverLang.Name), availableTestLabels) {
					out.w(`    - "//example/routeguide:%s_%s"`, clientLang.Name, serverLang.Name)
				}
			}
		}
	}

	//
	// Write tasks for examples
	//
	for _, lang := range languages {
		for _, rule := range lang.Rules {
			exampleDir := path.Join(dir, "example", lang.Dir, rule.Name)

			for _, ciPlatform := range ciPlatforms {
				if !doTestOnPlatform(lang, rule, ciPlatform) {
					continue
				}

				out.w("  %s_%s_%s:", lang.Name, rule.Name, ciPlatform)
				out.w("    name: '%s: %s'", lang.Name, rule.Name)
				out.w("    platform: %s", ciPlatform)
				if ciPlatform == "macos" {
					out.w("    build_flags:")
					out.w(`    - "--copt=-DGRPC_BAZEL_BUILD"`) // https://github.com/bazelbuild/bazel/issues/4341 required for macos
				}
				out.w("    build_targets:")
				out.w(`      - "//..."`)
				out.w("    working_directory: %s", exampleDir)

				if len(lang.PresubmitEnvVars) > 0 || len(rule.PresubmitEnvVars) > 0 {
					out.w("    environment:")
					for k, v := range lang.PresubmitEnvVars {
						out.w("      %s: %s", k, v)
					}
					for k, v := range rule.PresubmitEnvVars {
						out.w("      %s: %s", k, v)
					}
				}
			}
		}
	}

	// Add test workspaces
	for _, testWorkspace := range findTestWorkspaceNames(dir) {
		for _, ciPlatform := range ciPlatforms {
			if ciPlatform == "windows" && (testWorkspace == "python3_grpc" || testWorkspace == "python_deps") {
				continue // Don't run python grpc test workspaces on windows
			}
			out.w("  test_workspace_%s_%s:", testWorkspace, ciPlatform)
			out.w("    name: 'test workspace: %s'", testWorkspace)
			out.w("    platform: %s", ciPlatform)
			if ciPlatform == "macos" {
				out.w("    build_flags:")
				out.w(`    - "--copt=-DGRPC_BAZEL_BUILD"`) // https://github.com/bazelbuild/bazel/issues/4341 required for macos
			}
			out.w("    test_flags:")
			if ciPlatform == "macos" {
				out.w(`    - "--copt=-DGRPC_BAZEL_BUILD"`) // https://github.com/bazelbuild/bazel/issues/4341 required for macos
			}
			out.w(`    - "--test_output=errors"`)
			out.w("    test_targets:")
			out.w(`      - "//..."`)
			out.w("    working_directory: %s", path.Join(dir, "test_workspaces", testWorkspace))
		}
	}

	out.ln()
	out.MustWrite(filepath.Join(dir, ".bazelci", "presubmit.yml"))
}

func mustWriteExamplesMakefile(dir string, languages []*Language) {
	out := &LineWriter{}
	slashRegex := regexp.MustCompile("/")

	var allNames []string
	for _, lang := range languages {
		var langNames []string

		// Calculate depth of lang dir
		langDepth := len(slashRegex.FindAllStringIndex(lang.Dir, -1))

		// Create rules for each example
		for _, rule := range lang.Rules {
			exampleDir := path.Join(dir, "example", lang.Dir, rule.Name)

			var name = fmt.Sprintf("%s_%s_example", lang.Name, rule.Name)
			allNames = append(allNames, name)
			langNames = append(langNames, name)
			out.w(".PHONY: %s", name)
			out.w("%s:", name)
			out.w("	cd %s; \\", exampleDir)
			out.w("	bazel --batch build --verbose_failures --disk_cache=%s../../bazel-disk-cache //...", strings.Repeat("../", langDepth))
			out.ln()
		}

		// Create grouped rules for each language
		targetName := fmt.Sprintf("%s_examples", lang.Name)
		out.w(".PHONY: %s", targetName)
		out.w("%s: %s", targetName, strings.Join(langNames, " "))
		out.ln()
	}

	// Write all examples rule
	out.w(".PHONY: all_examples")
	out.w("all_examples: %s", strings.Join(allNames, " "))

	out.ln()
	out.MustWrite(filepath.Join(dir, "example", "Makefile.mk"))
}

func mustWriteTestWorkspacesMakefile(dir string) {
	out := &LineWriter{}

	// For each test workspace, add makefile rule
	var allNames []string
	for _, testWorkspace := range findTestWorkspaceNames(dir) {
		var name = fmt.Sprintf("test_workspace_%s", testWorkspace)
		allNames = append(allNames, name)
		out.w(".PHONY: %s", name)
		out.w("%s:", name)
		out.w("	cd %s; \\", path.Join(dir, "test_workspaces", testWorkspace))
		out.w("	bazel --batch test --verbose_failures --disk_cache=../bazel-disk-cache --test_output=errors //...")
		out.ln()
	}

	// Write all test workspaces rule
	out.w(".PHONY: all_test_workspaces")
	out.w("all_test_workspaces: %s", strings.Join(allNames, " "))

	out.ln()
	out.MustWrite(filepath.Join(dir, "test_workspaces", "Makefile.mk"))
}

func mustWriteHttpArchiveTestWorkspace(dir, ref, sha256 string) {
	out := &LineWriter{}
	out.w(`load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_proto_grpc",
    urls = ["https://github.com/rules-proto-grpc/rules_proto_grpc/archive/%s.tar.gz"],
    sha256 = "%s",
    strip_prefix = "rules_proto_grpc-%s",
)
`, ref, sha256, ref)
	out.MustWrite(filepath.Join(dir, "test_workspaces", "readme_http_archive", "WORKSPACE"))
}

func findTestWorkspaceNames(dir string) []string {
	files, err := ioutil.ReadDir(filepath.Join(dir, "test_workspaces"))
	if err != nil {
		log.Fatal(err)
	}

	var testWorkspaces []string
	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") && !strings.HasPrefix(file.Name(), "bazel-") {
			testWorkspaces = append(testWorkspaces, file.Name())
		}
	}

	return testWorkspaces
}
