package timeutil

import (
	"log"
	"testing"
	"time"
)

func TestTimeToEpoch(t *testing.T) {
	epoch := TimeToEpoch(time.Now())

	log.Printf("epoch: %+v", epoch)
}

func TestEpochToTime(t *testing.T) {
	tm := EpochToTime(1426730687)

	log.Printf("time: %+v", tm)
}

func TestParseTime(t *testing.T) {
	epoch := ParseTime("2015-03-27T00:00:00+08:00")

	log.Printf("time: %d", epoch)
}

func TestGetDateStringWithZone(t *testing.T) {
	epoch := ParseTime("2015-03-27T19:00:00+00:00")
	t1 := EpochToTime(epoch)

	dateString := GetDateStringWithZone(t1, "Asia/Taipei")

	log.Printf("date: %s", dateString)
}

func TestGetTimeStringWithZone(t *testing.T) {
	epoch := ParseTime("2015-03-27T19:00:00+00:00")
	t1 := EpochToTime(epoch)

	timeString := GetTimeStringWithZone(t1, "Asia/Taipei")

	log.Printf("time: %s", timeString)
}

func TestTimeTrack(t *testing.T) {
	defer TimeTrack(time.Now(), "TimeTrack")
	time.Sleep(time.Second)
}
