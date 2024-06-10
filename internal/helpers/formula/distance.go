package formula_helper

import (
	"math"
	"math/rand"
	"time"

	purchase_entity "github.com/danzBraham/beli-mang/internal/entities/purchase"
)

// Circle represents a circle by its center and radius
type Circle struct {
	Center purchase_entity.Location
	Radius float64
}

// distance returns the Haversine distance between two points
func Distance(p1, p2 purchase_entity.Location) float64 {
	const R = 6371 // Earth radius in kilometers
	lat1, lon1 := p1.Lat*math.Pi/180, p1.Long*math.Pi/180
	lat2, lon2 := p2.Lat*math.Pi/180, p2.Long*math.Pi/180
	dlat, dlon := lat2-lat1, lon2-lon1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// isInCircle checks if a point is inside a given circle
func isInCircle(p purchase_entity.Location, c Circle) bool {
	return Distance(p, c.Center) <= c.Radius
}

// circleFromTwoPoints returns a circle that passes through two points
func circleFromTwoPoints(p1, p2 purchase_entity.Location) Circle {
	center := purchase_entity.Location{
		Lat:  (p1.Lat + p2.Lat) / 2,
		Long: (p1.Long + p2.Long) / 2,
	}
	radius := Distance(p1, p2) / 2
	return Circle{Center: center, Radius: radius}
}

// circleFromThreePoints returns a circle that passes through three points
func circleFromThreePoints(p1, p2, p3 purchase_entity.Location) Circle {
	ax, ay := p1.Lat, p1.Long
	bx, by := p2.Lat, p2.Long
	cx, cy := p3.Lat, p3.Long

	d := 2 * (ax*(by-cy) + bx*(cy-ay) + cx*(ay-by))
	if d == 0 {
		return Circle{} // Collinear points, undefined circle
	}

	ux := ((ax*ax+ay*ay)*(by-cy) + (bx*bx+by*by)*(cy-ay) + (cx*cx+cy*cy)*(ay-by)) / d
	uy := ((ax*ax+ay*ay)*(cx-bx) + (bx*bx+by*by)*(ax-cx) + (cx*cx+cy*cy)*(bx-ax)) / d
	center := purchase_entity.Location{Lat: ux, Long: uy}
	radius := Distance(center, p1)
	return Circle{Center: center, Radius: radius}
}

// welzl recursively finds the smallest enclosing circle
func welzl(points []purchase_entity.Location, boundary []purchase_entity.Location, n int) Circle {
	if n == 0 || len(boundary) == 3 {
		return trivialCircle(boundary)
	}

	idx := rand.Intn(n)
	p := points[idx]
	points[idx], points[n-1] = points[n-1], points[idx]

	circle := welzl(points, boundary, n-1)

	if isInCircle(p, circle) {
		return circle
	}

	boundary = append(boundary, p)
	return welzl(points, boundary, n-1)
}

// trivialCircle finds the smallest circle from up to 3 points
func trivialCircle(points []purchase_entity.Location) Circle {
	switch len(points) {
	case 0:
		return Circle{Center: purchase_entity.Location{Lat: 0, Long: 0}, Radius: 0}
	case 1:
		return Circle{Center: points[0], Radius: 0}
	case 2:
		return circleFromTwoPoints(points[0], points[1])
	case 3:
		return circleFromThreePoints(points[0], points[1], points[2])
	}
	return Circle{}
}

// SmallestEnclosingCircle returns the smallest enclosing circle for a set of points
func SmallestEnclosingCircle(points []purchase_entity.Location) Circle {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := make([]purchase_entity.Location, len(points))
	copy(shuffled, points)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return welzl(shuffled, nil, len(shuffled))
}
