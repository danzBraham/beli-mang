package purchase_exception

import "errors"

var (
	ErrDistanceTooFar     = errors.New("the distance is too far")
	ErrEstimateIdNotFound = errors.New("estimate id is not found")
)
