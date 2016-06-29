package sync

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm"

	"straas.io/external/rdbms"
	"straas.io/pierce"
)

// newSinker creates a sinker
func newSinker(db *gorm.DB) (sinker, error) {
	s := &gormSinker{
		db: db,
	}
	if err := s.ensureSchema(); err != nil {
		return nil, err
	}
	return s, nil
}

// TODO: leverge namedQueue for better performance
type gormSinker struct {
	db *gorm.DB
}

func (g *gormSinker) ensureSchema() error {
	errs := g.db.
		AutoMigrate(Record{}).
		AddUniqueIndex("ux_cluster_version", "cluster", "version").
		GetErrors()
	return combineErrors(errs)
}

func (g *gormSinker) Sink(roomMeta pierce.RoomMeta, data interface{}, version uint64) error {
	v, _ := json.Marshal(data)

	rec := Record{
		Namespace: roomMeta.Namespace,
		Room:      roomMeta.ID,
		Value:     string(v),
		Cluster:   1, // hard-coded is enough now
		Version:   version,
	}
	errs := g.db.Create(&rec).GetErrors()

	if len(errs) == 1 && rdbms.IsErrConstraintUnique(errs[0]) {
		return nil
	}
	return combineErrors(errs)
}

func (g *gormSinker) Diff(namespace string, index uint64, size int) ([]Record, error) {
	var rec []Record
	errs := g.db.
		Where("ID > ? AND Namespace = ?", index, namespace).
		Limit(size).
		Find(&rec).
		GetErrors()

	return rec, combineErrors(errs)
}

type result struct {
	Version uint64
}

func (g *gormSinker) LatestVersion() (uint64, error) {
	var r result
	errs := g.db.Table("records").
		Select("max(version) as version").
		// hard-coded is enough now
		Where("cluster = ?", 1).
		Scan(&r).
		GetErrors()
	err := combineErrors(errs)
	if err != nil {
		return 0, err
	}
	return r.Version, nil
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
