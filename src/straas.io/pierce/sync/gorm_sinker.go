package sync

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"straas.io/external/rdbms"
	"straas.io/pierce"
)

func NewSinker(db *gorm.DB, namespace string) *GormSinker {
	return &GormSinker{
		db:        db,
		namespace: namespace,
	}
}

type GormSinker struct {
	db        *gorm.DB
	namespace string
}

// leverge namedQueue for better performance
func (g *GormSinker) Sink(roomMeta pierce.RoomMeta, data interface{}, version uint64) error {
	v, _ := json.Marshal(data)

	// TODO: marshall as string
	rec := Record{
		Namespace: roomMeta.Namespace,
		Room:      roomMeta.ID,
		Value:     string(v),
		Cluster:   1, // hard-code is enough now
		Version:   version,
	}
	errs := g.db.Create(&rec).GetErrors()

	switch {
	case len(errs) == 0:
		return nil

	// violating unique constrain means such data already exist
	case len(errs) == 1 && rdbms.IsErrConstraintUnique(errs[0]):
		return nil

	default:
		return combineErrors(errs)
	}
}

func (g *GormSinker) Diff(index uint64, size int) ([]Record, error) {
	var rec []Record
	errs := g.db.
		Where("ID > ? AND Namespace = ?", index, g.namespace).
		Limit(size).
		Find(&rec).
		GetErrors()

	return rec, combineErrors(errs)
}

type Result struct {
	Version uint64
}

func (g *GormSinker) LatestVersion() (uint64, error) {
	var result Result
	errs := g.db.Table("records").
		Select("max(version) as version").
		Where("namespace = ?", g.namespace).
		Scan(result).
		GetErrors()
	err := combineErrors(errs)
	if err != nil {
		return 0, err
	}
	return result.Version, nil
}

func combineErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[1]
	}
	// multiple erros, combine them
	buf := bytes.NewBuffer(nil)
	for i, err := range errs {
		if i > 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(err.Error())
	}
	return errors.New(buf.String())
}
