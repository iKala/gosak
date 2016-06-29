package sync

import (
	"testing"

	"github.com/jinzhu/gorm"
	// register sqlite driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/suite"

	"straas.io/pierce"
)

func TestGormSinker(t *testing.T) {
	suite.Run(t, new(gormSinkerTestSuite))
}

type gormSinkerTestSuite struct {
	suite.Suite
	db   *gorm.DB
	impl *gormSinker
}

func (s *gormSinkerTestSuite) SetupTest() {
	var err error
	// create in memory sqlite for test
	s.db, err = gorm.Open("sqlite3", "file::memory:")
	s.NoError(err)

	sinker, err := newSinker(s.db)
	s.NoError(err)
	s.impl = sinker.(*gormSinker)
}

func (s *gormSinkerTestSuite) TestSinkDiffVersion() {
	err := s.impl.Sink(pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        "aaa",
	}, "1234", 10)
	s.NoError(err)

	err = s.impl.Sink(pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        "aaa",
	}, "1235", 12)
	s.NoError(err)

	err = s.impl.Sink(pierce.RoomMeta{
		Namespace: testNamespace2,
		ID:        "aaa",
	}, "1236", 14)
	s.NoError(err)

	var recs []Record
	errs := s.db.Find(&recs).GetErrors()
	s.Equal(len(errs), 0)
	s.Equal(len(recs), 3)

	s.Equal(recs, []Record{
		Record{
			ID:        1,
			Namespace: testNamespace,
			Room:      "aaa",
			Value:     `"1234"`,
			Version:   10,
			Cluster:   1,
			// unable to mock time
			CreatedAt: recs[0].CreatedAt,
		},
		Record{
			ID:        2,
			Namespace: testNamespace,
			Room:      "aaa",
			Value:     `"1235"`,
			Version:   12,
			Cluster:   1,
			CreatedAt: recs[1].CreatedAt,
		},
		// different namespace
		Record{
			ID:        3,
			Namespace: testNamespace2,
			Room:      "aaa",
			Value:     `"1236"`,
			Version:   14,
			Cluster:   1,
			CreatedAt: recs[2].CreatedAt,
		},
	})

	version, err := s.impl.LatestVersion()
	s.NoError(err)
	s.Equal(version, uint64(14))

	recs, err = s.impl.Diff(testNamespace, 1, 5)
	s.NoError(err)
	s.Equal(recs, []Record{
		Record{
			ID:        2,
			Namespace: testNamespace,
			Room:      "aaa",
			Value:     `"1235"`,
			Version:   12,
			Cluster:   1,
			CreatedAt: recs[0].CreatedAt,
		},
	})
}

func (s *gormSinkerTestSuite) TestSinkUnique() {
	err := s.impl.Sink(pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        "aaa",
	}, "1234", 10)
	s.NoError(err)

	// second one shoule be ignored.
	err = s.impl.Sink(pierce.RoomMeta{
		Namespace: testNamespace,
		ID:        "aaa",
	}, "1235", 10)
	s.NoError(err)

	var recs []Record
	errs := s.db.Find(&recs).GetErrors()
	s.Equal(len(errs), 0)
	s.Equal(len(recs), 1)

	s.Equal(recs, []Record{
		Record{
			ID:        1,
			Namespace: testNamespace,
			Room:      "aaa",
			Value:     `"1234"`,
			Version:   10,
			Cluster:   1,
			// unable to mock time
			CreatedAt: recs[0].CreatedAt,
		},
	})
}
