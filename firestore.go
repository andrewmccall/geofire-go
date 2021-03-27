package geofire

import (
	firestore "cloud.google.com/go/firestore"
	"google.golang.org/genproto/googleapis/type/latlng"
)

// GeoWhere runs a geo query against the collection.
func GeoWhere(ref *firestore.CollectionRef, location *latlng.LatLng, radius uint) {
	queries := QueryiesAtLocation(location, float64(radius))

}
