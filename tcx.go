package tcx

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

// Tcx represents the root of a TCX file
type Tcx struct {
	XMLName      xml.Name   `xml:"TrainingCenterDatabase"`
	XMLNs        string     `xml:"xmlns,attr"`
	XMLNsXsi     string     `xml:"xsi,attr,omitempty"`
	XMLNsXsd     string     `xml:"xsd,attr,omitempty"`
	XMLSchemaLoc string     `xml:"schemaLocation,attr,omitempty"`
	Activities   []Activity `xml:"Activities>Activity"`
}

type Activity struct {
	Sport   string    `xml:"Sport,attr"`
	ID      time.Time `xml:"Id"`
	Creator Creator   `xml:"Creator"`
	Laps    []Lap     `xml:"Lap"`
}

type Creator struct {
	Name      string `xml:"Name"`
	UnitID    int    `xml:"UnitId"`
	ProductID int    `xml:"ProductID"`
}

type Lap struct {
	StartTime                  time.Time    `xml:"StartTime,attr"`
	TotalTimeInSeconds         float64      `xml:"TotalTimeSeconds"`
	DistanceInMeters           float64      `xml:"DistanceMeters"`
	MaximumSpeedInMetersPerSec float64      `xml:"MaximumSpeed"`
	Calories                   float64      `xml:"Calories"`
	Intensity                  string       `xml:"Intensity"`
	TriggerMethod              string       `xml:"TriggerMethod"`
	Track                      []Trackpoint `xml:"Track>Trackpoint"`
}

type Trackpoint struct {
	Time                time.Time `xml:"Time"`
	LatitudeInDegrees   float64   `xml:"LatitudeDegrees"`
	LongitudeInDegrees  float64   `xml:"LongitudeDegrees"`
	AltitudeInMeters    float64   `xml:"AltitudeMeters"`
	HeartRateInBpm      int       `xml:"HeartRateBpm>Value"`
	Cadence             int       `xml:"Cadence"`
	SpeedInMetersPerSec float64   `xml:"Extensions>TPX>Speed"`
}

type Pace struct {
	float64
}

// Parse parses a TCX reader and return a Tcx object.
func Parse(r io.Reader) (*Tcx, error) {
	g := NewTcx()
	d := xml.NewDecoder(r)
	err := d.Decode(g)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse tcx data: %v", err)
	}
	return g, nil
}

// ParseFile reads a TCX file and parses it.
func ParseFile(filepath string) (*Tcx, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// NewTcx creates and returns a new Gpx objects.
func NewTcx() *Tcx {
	tcx := new(Tcx)
	return tcx
}

//
// Reports the start time in terms of the local time.
//
func (a *Activity) StartTime() time.Time {
	return a.ID.In(time.Local)
}

func (a *Activity) TotalDuration() time.Duration {
	var duration time.Duration = 0
	for _, l := range a.Laps {
		duration += time.Duration(l.TotalTimeInSeconds) * time.Second
	}
	return duration
}

func (a *Activity) TotalDistance() float64 {
	var d float64 = 0
	for _, l := range a.Laps {
		d += l.DistanceInMeters
	}
	return d
}

//
// Converts meters to miles.
//
func (a *Activity) TotalDistanceInMiles() float64 {
	return a.TotalDistance() * 0.00062137
}

func (a *Activity) AverageHeartbeat() float64 {
	var totalhr int = 0
	var nbhr int = 0
	for _, l := range a.Laps {
		for _, p := range l.Track {
			totalhr += p.HeartRateInBpm
			nbhr += 1
		}
	}
	return float64(totalhr) / float64(nbhr)
}

func (p *Pace) String() string {
	intpart, fracpart := math.Modf(p.float64)
	return fmt.Sprintf("%.f:%.f", intpart, fracpart*60)
}

func GetPaceFromSpeedInMs(speed float64) *Pace {
	var p *Pace = new(Pace)
	p.float64 = 50 / (speed * 3)
	return p
}

func (a *Activity) AveragePace() *Pace {
	var totals float64 = 0
	var nbs int = 0
	for _, l := range a.Laps {
		for _, p := range l.Track {
			totals += p.SpeedInMetersPerSec
			nbs += 1
		}
	}
	return GetPaceFromSpeedInMs(totals / float64(nbs))
}
