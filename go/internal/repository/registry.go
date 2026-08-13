package repository

import "fmt"

type Connector string

const (
	ConnectorVisualStudioMarketplace Connector = "visual-studio-marketplace"
	ConnectorOpenVSX                 Connector = "open-vsx"
	ConnectorCursorMarketplace       Connector = "cursor-marketplace"
	ConnectorWindsurfMarketplace     Connector = "windsurf-marketplace"
)

type Definition struct {
	ID        string
	Connector Connector
}

var builtIns = map[string]Definition{
	"visual-studio-marketplace": {ID: "visual-studio-marketplace", Connector: ConnectorVisualStudioMarketplace},
	"open-vsx":                  {ID: "open-vsx", Connector: ConnectorOpenVSX},
	"cursor-marketplace":        {ID: "cursor-marketplace", Connector: ConnectorCursorMarketplace},
	"windsurf-marketplace":      {ID: "windsurf-marketplace", Connector: ConnectorWindsurfMarketplace},
}

func Lookup(id string) (Definition, error) {
	definition, ok := builtIns[id]
	if !ok {
		return Definition{}, fmt.Errorf("Extension repository is not configured: %s", id)
	}
	return definition, nil
}
