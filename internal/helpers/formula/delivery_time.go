package formula_helper

import (
	"math"
)

// CalculateDeliveryTime calculates the estimated delivery time in minutes
// based on the maximum distance in kilometers and an assumed average speed.
func CalculateDeliveryTime(maxDistance float64) int {
	const averageSpeed = 40.0 // Average speed in km/h

	// Delivery time in hours
	deliveryTimeHours := maxDistance / averageSpeed

	// Convert delivery time to minutes and round to the nearest integer
	deliveryTimeMinutes := int(math.Round(deliveryTimeHours * 60))

	return deliveryTimeMinutes
}
