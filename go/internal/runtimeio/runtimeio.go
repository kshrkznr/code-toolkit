package runtimeio

import (
	"context"
	"errors"

	"code-toolkit/internal/cookbook"
	"code-toolkit/internal/runtimeartifact"
	"code-toolkit/internal/settings"
)

var ErrUnsupported = errors.New("Runtime Artifact unsupported")

type Scope struct{ Name string }

type Extension struct {
	ID      string
	Version string
}

func DefaultScope() Scope            { return Scope{} }
func ProfileScope(name string) Scope { return Scope{Name: name} }
func (s Scope) IsDefault() bool      { return s.Name == "" }

type Runtime interface {
	Scopes(context.Context) ([]Scope, error)
	EnsureProfile(context.Context, string) error
	SetInheritance(context.Context, Scope, cookbook.Inheritance) error
	ReadInheritance(context.Context, Scope) (cookbook.Inheritance, error)
	ReadSettings(context.Context, Scope) (settings.Document, error)
	WriteSettings(context.Context, Scope, settings.Document) error
	Extensions(context.Context, Scope) ([]Extension, error)
	InstallExtension(context.Context, Scope, string) error
	UninstallExtension(context.Context, Scope, string) error
}

type ArtifactRuntime interface {
	ReadKeybindings(context.Context, Scope) (runtimeartifact.Array, error)
	WriteKeybindings(context.Context, Scope, runtimeartifact.Array) error
	ReadTasks(context.Context, Scope) (runtimeartifact.Object, error)
	WriteTasks(context.Context, Scope, runtimeartifact.Object) error
	ReadMCP(context.Context, Scope) (runtimeartifact.Object, error)
	WriteMCP(context.Context, Scope, runtimeartifact.Object) error
	ReadSnippets(context.Context, Scope) (runtimeartifact.Snippets, error)
	WriteSnippets(context.Context, Scope, runtimeartifact.Snippets) error
}
