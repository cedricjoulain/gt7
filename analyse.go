package main

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"
)

type Message [296]byte

func (m Message) RPM() float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(m[15*4:]))
}

type XYZ struct {
	X float32 // 0:4
	Y float32 //4:8
	Z float32 //8:12
}

// String, comma separated
func (v XYZ) String() string {
	return fmt.Sprintf("%f,%f,%f", v.X, v.Y, v.Z)
}

// Norm, Euclidean norm separated
func (v XYZ) Norm() float64 {
	return math.Sqrt(float64(v.X)*float64(v.X) + float64(v.Y)*float64(v.Y) + float64(v.Z)*float64(v.Z))
}

// Header, comma separated
func XYZHeader(name string) string {
	return fmt.Sprintf("%s X, %s Y, %s Z", name, name, name)
}

// TimeSpan number of elapsed milliseconds
type TimeSpan int32

func (t TimeSpan) String() string {
	return (time.Duration(t) * time.Duration(time.Millisecond)).String()
}

// SimulatorFlags Flags/States of the simulation.
type SimulatorFlags int16

const (
	None SimulatorFlags = 0
	/// The car is on the track or paddock, with data available.
	CarOnTrack SimulatorFlags = 1 << iota
	/// The game's simulation is paused.
	/// Note: The simulation will not be paused while in the pause menu in online modes.
	Paused

	/// Track or car is currently being loaded onto the track.
	LoadingOrProcessing

	/// Needs more investigation
	InGear

	/// Current car has a Turbo.
	HasTurbo

	/// Rev Limiting is active.
	RevLimiterBlinkAlertActive

	/// Hand Brake is active.
	HandBrakeActive

	/// Lights are active.
	LightsActive

	/// High Beams are turned on.
	HighBeamActive

	/// Low Beams are turned on.
	LowBeamActive

	/// ASM is active.
	ASMActive

	/// Traction Control is active.
	TCSActive
)

type Packet struct {
	Magic                          int32          // 0:4
	Position                       XYZ            // 4:16
	Velocity                       XYZ            // 16:28
	Rotation                       XYZ            // 28:40
	RelativeOrientationToNorth     float32        // 40:44
	AngularVelocity                XYZ            // 44:56
	BodyHeight                     float32        // 56:60
	EngineRPM                      float32        // 60:64
	Reserved1                      int32          // 64:68
	GasLevel                       float32        // 68:74
	GasCapacity                    float32        // 74:78
	MetersPerSecond                float32        // 78:82
	TurboBoost                     float32        // 82:86
	OilPressure                    float32        // 86:90
	WaterTemperature               float32        // 90:94
	OilTemperature                 float32        // 94:98
	TireFL_SurfaceTemperature      float32        // 102:106
	TireFR_SurfaceTemperature      float32        // 106:110
	TireRL_SurfaceTemperature      float32        // 110:114
	TireRR_SurfaceTemperature      float32        // 114:118
	PacketId                       int32          // 118:122
	LapCount                       int16          // 122:124
	LapsInRace                     int16          // 124:126
	BestLapTime                    TimeSpan       // 126:130
	LastLapTime                    TimeSpan       // 130:134
	TimeOfDayProgression           TimeSpan       // 134:138
	PreRaceStartPositionOrQualiPos int16          // 138:140
	NumCarsAtPreRace               int16          // 140:142
	MinAlertRPM                    int16          // 142:144
	MaxAlertRPM                    int16          // 144:148
	CalculatedMaxSpeed             int16          // 148:150
	Flags                          SimulatorFlags // 150:152

	// CurrentGear = (byte)(bits & 0b1111); // 4 bits
	// SuggestedGear = (byte)(bits >> 4); // Also 4 bits
	Gear       byte // 152:153
	Throttle   byte // 153:154
	Brake      byte // 154:155
	Empty_0x93 byte // 155:156

	RoadPlane         XYZ     // 156:168
	RoadPlaneDistance float32 // 168:172

	WheelFL_RevPerSecond float32 // 172:176
	WheelFR_RevPerSecond float32 // 176:188
	WheelRL_RevPerSecond float32 // 180:184
	WheelRR_RevPerSecond float32 // 184:188
	TireFL_TireRadius    float32 // 188:192
	TireFR_TireRadius    float32 // 192:196
	TireRL_TireRadius    float32 // 196:200
	TireRR_TireRadius    float32 // 200:204
	TireFL_SusHeight     float32 // 204:208
	TireFR_SusHeight     float32 // 208:212
	TireRL_SusHeight     float32 // 212:216
	TireRR_SusHeight     float32 // 216:220

	// Seems to be reserved - server does not set that
	Reserved2 [8]int32 // 220:252

	ClutchPedal            float32 // 252:256
	ClutchEngagement       float32 // 256:260
	RPMFromClutchToGearbox float32 // 260:264

	TransmissionTopSpeed float32 // 264:268

	// Always read as a fixed 7 gears
	// Normally 8th is not set at all. The game memcpy's the gear ratios without bound checking
	// The LC500 which has 10 gears even overrides the car code
	GearRatios [8]float32 // 268:300

	CarCode int32 // 300:304
}

// EchartsData generate data string (json) for echarts
func EchartsData(packets []Packet, data *[][]float64) {
	*data = make([][]float64, len(packets))
	for i, p := range packets {
		(*data)[i] = []float64{
			float64(p.Position.X),
			float64(-p.Position.Z),
			float64(p.Brake),
			float64(p.Throttle),
		}
	}
	return
}

// CsvLine dump one line of information comma separated
// Order must respect CsvHeader
func (p Packet) CsvLine() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s,", p.TimeOfDayProgression)
	fmt.Fprintf(&b, "%d,", p.LapCount)
	fmt.Fprintf(&b, "%s,", p.BestLapTime)
	fmt.Fprintf(&b, "%d,", p.Throttle)
	fmt.Fprintf(&b, "%d,", p.Brake)
	fmt.Fprintf(&b, "%f,", p.MetersPerSecond)
	fmt.Fprintf(&b, "%s,", p.Position)
	fmt.Fprintf(&b, "%s,", p.RoadPlane)
	fmt.Fprintf(&b, "%f,", p.RoadPlaneDistance)
	return b.String()
}

// CsvHeader dump header informatio comma separated
// Order must respect CsvLine
func (p Packet) CsvHeader() string {
	var b strings.Builder
	fmt.Fprintf(&b, "TimeOfDayProgression,")
	fmt.Fprintf(&b, "LapCount,")
	fmt.Fprintf(&b, "BestLapTime,")
	fmt.Fprintf(&b, "Throttle,")
	fmt.Fprintf(&b, "Brake,")
	fmt.Fprintf(&b, "MetersPerSecond,")
	fmt.Fprintf(&b, "%s,", XYZHeader("Position"))
	fmt.Fprintf(&b, "%s,", XYZHeader("RoadPlane"))
	fmt.Fprintf(&b, "RoadPlaneDistance")
	return b.String()
}

// Analyse open raw decoded data and parse
func Analyse(filename string, lap int) (packets []Packet, err error) {
	var (
		nbr    int
		r      io.Reader
		f      *os.File
		header [2]byte
		packet Packet
	)
	if f, err = os.Open(filename); err != nil {
		err = fmt.Errorf("unable to open %s:%s", filename, err)
		return
	}
	defer f.Close()
	// is it a gzip file ?
	// use magic number
	if _, err = io.ReadFull(f, header[:]); err != nil {
		err = fmt.Errorf("unable to read head os %s:%s", filename, err)
		return
	}
	// back to beginning of file
	if _, err = f.Seek(0, 0); err != nil {
		return
	}
	if header[0] == 0x1f && header[1] == 0x8b {
		// this is a gzipRPM
		if r, err = gzip.NewReader(f); err != nil {
			err = fmt.Errorf("unable to create gzip reader for %s:%s", filename, err)
			return
		}
	} else {
		r = f
	}
	prevLap := -1
	fmt.Println(packet.CsvHeader())
	packets = make([]Packet, 0)
	for err == nil {
		if err = binary.Read(r, binary.LittleEndian, &packet); err != nil {
			break
		}
		nbr++
		if packet.LapCount != int16(prevLap) && packet.BestLapTime != -1 {
			// ending lap with time
			fmt.Println(packet.CsvLine())
		}
		prevLap = int(packet.LapCount)
		if packet.LapCount == int16(lap) {
			packets = append(packets, packet)
		}
	}
	if err == io.EOF {
		// just the end
		err = nil
	}
	return
}
