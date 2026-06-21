package domain

import "errors"

// Feature-level sentinel errors, shared between repository and usecase so the
// outer (repo) layer can surface not-found without the inner (usecase) layer
// needing to map strings.
var (
	ErrVehicleTypeNotFound = errors.New("vehicle type not found")
)
