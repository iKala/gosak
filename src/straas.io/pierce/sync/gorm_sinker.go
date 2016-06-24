package sync

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"straas.io/pierce"
)

type GormSinker struct {
	db *gorm.DB
}

func (g *GormSinker) Sink(roomMeta pierce.RoomMeta, data interface{}, version uint64) error {
	// TODO: marshall as string
	rec := Record{
		Namespace: roomMeta.Namespace,
		Room:      roomMeta.ID,
		Value:     "",
		Cluster:   1,
		Version:   version,
	}

	errs := g.db.Create(&rec).GetErrors()
	if len(errs) == 0 {
		return nil
	}
	return nil
}

func (g *GormSinker) Diff(index uint64, size int) ([]Record, error) {
	return nil, nil
}
