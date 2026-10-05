package checkpoint

import (
	"context"
	"fmt"

	"github.com/warpstreamlabs/bento/public/service"
)

// Store persists opaque per-stream cursors; Get returns nil when a key has never been set.
type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Close(ctx context.Context) error
}

const (
	fieldType = "type"
	fieldPath = "path"
	typeFile  = "file"
)

func ConfigField(name string) *service.ConfigField {
	return service.NewObjectField(name,
		service.NewStringEnumField(fieldType, typeFile).
			Description("Where cursors are persisted.").
			Default(typeFile),
		service.NewStringField(fieldPath).
			Description("File path for the `file` store; must be on a persistent volume.").
			Default("./state/checkpoints.json"),
	).Description("Cursor persistence. A cursor only advances after its events are acknowledged by the output.")
}

func FromConfig(conf *service.ParsedConfig, name string) (Store, error) {
	ns := conf.Namespace(name)
	kind, err := ns.FieldString(fieldType)
	if err != nil {
		return nil, err
	}
	switch kind {
	case typeFile:
		path, err := ns.FieldString(fieldPath)
		if err != nil {
			return nil, err
		}
		return OpenFile(path)
	default:
		return nil, fmt.Errorf("unsupported checkpoint type %q", kind)
	}
}
