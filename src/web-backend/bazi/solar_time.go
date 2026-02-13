package bazi

import (
	"fmt"
	"math"
	"time"
)

// TrueSolarTimeFromBeijing converts a Beijing standard time (UTC+08:00, based on 120E)
// to local true (apparent) solar time for a given location longitude.
//
// Inputs:
// - solarDate: "YYYY-MM-DD"
// - birthTime: "HH:mm"
//
// Output:
// - "YYYY/MM/DD HH:mm"
//
// Model:
//
//	true_solar_time = beijing_time + 4*(lon-120) minutes + EoT(date)
//
// where EoT is the equation of time in minutes (approximation).
func TrueSolarTimeFromBeijing(solarDate, birthTime string, longitudeDeg float64) (string, error) {
	if solarDate == "" || birthTime == "" {
		return "", fmt.Errorf("empty solarDate/birthTime")
	}
	if math.IsNaN(longitudeDeg) || math.IsInf(longitudeDeg, 0) {
		return "", fmt.Errorf("invalid longitude")
	}

	// Avoid depending on system tzdata. Beijing time is fixed UTC+08:00.
	bj := time.FixedZone("CST", 8*3600)
	t, err := time.ParseInLocation("2006-01-02 15:04", solarDate+" "+birthTime, bj)
	if err != nil {
		return "", err
	}

	// Longitude correction relative to Beijing standard meridian (120E).
	lonMinutes := 4.0 * (longitudeDeg - 120.0)
	eotMinutes := equationOfTimeMinutes(t)
	corrMinutes := lonMinutes + eotMinutes

	// Apply correction with rounding to the nearest second.
	sec := int64(math.Round(corrMinutes * 60.0))
	out := t.Add(time.Duration(sec) * time.Second)

	return out.In(bj).Format("2006/01/02 15:04"), nil
}

// equationOfTimeMinutes returns the equation of time in minutes for the given date.
// Approximation formula with typical error on the order of ~<1 minute.
func equationOfTimeMinutes(t time.Time) float64 {
	// Use day-of-year for the given local date (Beijing date is fine for the purposes here).
	n := float64(t.YearDay())
	b := 2 * math.Pi * (n - 81) / 364.0
	// Minutes.
	return 9.87*math.Sin(2*b) - 7.53*math.Cos(b) - 1.5*math.Sin(b)
}
