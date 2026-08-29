package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kshrkznr/code-toolkit/go/internal/docbundle"
)

func main() {
	repositoryRoot := flag.String("root", "..", "CTK repository root")
	output := flag.String("output", "", "output Bundle ZIP path")
	version := flag.String("version", "dev", "CTK version recorded in the Manifest")
	revision := flag.String("revision", "unknown", "source revision recorded in the Manifest")
	tag := flag.String("tag", "", "optional Release tag recorded in the Manifest")
	flag.Parse()
	if flag.NArg() != 0 || *output == "" {
		fmt.Fprintln(os.Stderr, "usage: ctk-docbundle -output <bundle.zip> [-root <repository>] [-version <version>] [-revision <commit>] [-tag <tag>]")
		os.Exit(2)
	}
	result, err := docbundle.Generate(*repositoryRoot, docbundle.Metadata{Version: *version, Revision: *revision, Tag: *tag})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, result.Archive, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s\ncontent-sha256: %s\n", *output, result.Manifest.ContentSHA256)
}
