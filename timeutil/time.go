package timeutil

import (
	"fmt"
	"log"
	"time"
)

// TimeToEpoch get epoch time from time
func TimeToEpoch(t time.Time) int {
	return int(t.Unix())
}

// EpochToTime get epoch time from time
func EpochToTime(e int) time.Time {
	tm := time.Unix(int64(e), 0)

	return tm
}

// ParseTime get epoch time from rfc3339 format time
func ParseTime(rfc3339Time string) int {
	t, _ := time.Parse(
		time.RFC3339,
		rfc3339Time)

	return TimeToEpoch(t)
}

// GetDateStringWithZone returns formatted date string
func GetDateStringWithZone(t time.Time, location string) string {
	loc, _ := time.LoadLocation(location)

	year, month, day := t.In(loc).Date()

	return fmt.Sprintf("%d-%d-%d", year, month, day)
}

// GetTimeStringWithZone returns formatted time string
func GetTimeStringWithZone(t time.Time, location string) string {
	loc, _ := time.LoadLocation(location)

	localTime := t.In(loc)

	year, month, day := localTime.Date()

	return fmt.Sprintf("%d/%d/%d %d:%d:%d",
		month, day, year, localTime.Hour(), localTime.Minute(), localTime.Second())
}

// TimeTrack measures execution time of method
// For example:
//
//    func foo() {
//      defer TimeTrack(time.Now(), "foo")
//      // Do something
//    }
//
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
