package formula_helper

import "math"

const EarthRadius = 6371.0

func Haversine(lat1, lat2, long1, long2 float64) float64 {
	dlat := (lat2 - lat1) * (math.Pi / 180.0)
	dlong := (long2 - long1) * (math.Pi / 180.0)

	lat1 = lat1 * (math.Pi / 180.0)
	lat2 = lat2 * (math.Pi / 180.0)

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Sin(dlong/2)*math.Sin(dlong/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}
