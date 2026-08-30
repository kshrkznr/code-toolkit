package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kshrkznr/code-toolkit/go/internal/docbundle"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, outputWriter io.Writer) error {
	flags := flag.NewFlagSet("ctk-docbundle", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	repositoryRoot := flags.String("root", "..", "CTK repository root")
	output := flags.String("output", "", "generated Bundle ZIP path")
	input := flags.String("input", "", "existing Bundle ZIP to validate and append")
	appendTo := flags.String("append-to", "", "executable to receive a validated Bundle")
	verifyExecutable := flags.String("verify-executable", "", "executable containing a Bundle to verify")
	version := flags.String("version", "dev", "CTK version recorded in a generated Manifest")
	revision := flags.String("revision", "unknown", "source revision recorded in a generated Manifest")
	tag := flags.String("tag", "", "Release tag recorded in a generated Manifest")
	expectVersion := flags.String("expect-version", "", "required Manifest CTK version")
	expectRevision := flags.String("expect-revision", "", "required Manifest source revision")
	expectTag := flags.String("expect-tag", "", "required Manifest Release tag")
	expectContent := flags.String("expect-content-sha256", "", "required Manifest aggregate content digest")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return usageError()
	}
	modes := 0
	for _, selected := range []bool{*output != "", *input != "", *verifyExecutable != ""} {
		if selected {
			modes++
		}
	}
	if modes != 1 || (*verifyExecutable != "" && *appendTo != "") {
		return usageError()
	}

	if *verifyExecutable != "" {
		bundle, err := docbundle.OpenExecutable(*verifyExecutable)
		if err != nil {
			return err
		}
		if err := verifyExpected(bundle.Manifest(), *expectVersion, *expectRevision, *expectTag, *expectContent); err != nil {
			return err
		}
		printManifest(outputWriter, bundle.Manifest())
		return nil
	}

	if *input != "" {
		archive, err := os.ReadFile(*input)
		if err != nil {
			return fmt.Errorf("read Documentation Bundle input: %w", err)
		}
		bundle, err := docbundle.Open(archive)
		if err != nil {
			return err
		}
		if err := verifyExpected(bundle.Manifest(), *expectVersion, *expectRevision, *expectTag, *expectContent); err != nil {
			return err
		}
		if *appendTo != "" {
			if err := docbundle.AppendExecutable(*appendTo, archive); err != nil {
				return err
			}
			fmt.Fprintf(outputWriter, "appended: %s\n", *appendTo)
		}
		printManifest(outputWriter, bundle.Manifest())
		return nil
	}

	result, err := docbundle.Generate(*repositoryRoot, docbundle.Metadata{Version: *version, Revision: *revision, Tag: *tag})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(*output, result.Archive, 0o644); err != nil {
		return err
	}
	if *appendTo != "" {
		if err := docbundle.AppendExecutable(*appendTo, result.Archive); err != nil {
			return err
		}
	}
	fmt.Fprintf(outputWriter, "%s\n", *output)
	printManifest(outputWriter, result.Manifest)
	return nil
}

func verifyExpected(manifest docbundle.Manifest, version, revision, tag, content string) error {
	for _, expectation := range []struct {
		name string
		want string
		got  string
	}{
		{name: "ctk-version", want: version, got: manifest.Version},
		{name: "source-revision", want: revision, got: manifest.Revision},
		{name: "release-tag", want: tag, got: manifest.Tag},
		{name: "content-sha256", want: content, got: manifest.ContentSHA256},
	} {
		if expectation.want != "" && expectation.got != expectation.want {
			return fmt.Errorf("Documentation Bundle %s mismatch: got %s, want %s", expectation.name, expectation.got, expectation.want)
		}
	}
	return nil
}

func printManifest(output io.Writer, manifest docbundle.Manifest) {
	fmt.Fprintf(output, "ctk-version: %s\n", manifest.Version)
	fmt.Fprintf(output, "source-revision: %s\n", manifest.Revision)
	fmt.Fprintf(output, "release-tag: %s\n", manifest.Tag)
	fmt.Fprintf(output, "definition-sha256: %s\n", manifest.DefinitionSHA256)
	fmt.Fprintf(output, "content-sha256: %s\n", manifest.ContentSHA256)
}

func usageError() error {
	return fmt.Errorf("usage: ctk-docbundle -output <bundle.zip> [-root <repository>] [-version <version>] [-revision <commit>] [-tag <tag>] [-append-to <executable>] | -input <bundle.zip> [-append-to <executable>] [-expect-*] | -verify-executable <executable> [-expect-*]")
}
