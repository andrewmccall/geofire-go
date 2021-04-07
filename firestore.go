package geofire

import (
	"context"
	"log"

	firestore "cloud.google.com/go/firestore"
	"github.com/andrewmccall/geoutils"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/type/latlng"
)

type GeoDocument struct {
	GeoHash  string         `firestore:"g,omitempty"`
	Location *latlng.LatLng `firestore:"l,omitempty"`
}

func (g *GeoDocument) SetLocation(loc *latlng.LatLng) {
	g.Location = loc
	g.GeoHash, _ = geoutils.GeoHash(loc.Latitude, loc.Longitude, DEFAULT_PRECISION)
}

type GeoSearchResult struct {
	GeoDocument
	Distance uint
	Doc      *firestore.DocumentSnapshot
}

// GeoWhere runs a geo query against the collection.
func GeoWhere(ref *firestore.CollectionRef, location *latlng.LatLng, radius uint, ctx context.Context) *GeoDocumentIterator {

	queries := QueryiesAtLocation(location, float64(radius))

	var iterators []*firestore.DocumentIterator
	for query := range queries {
		q := ref.OrderBy("g", firestore.Asc).StartAt(query.StartValue).EndAt(query.EndValue).Documents(ctx)
		iterators = append(iterators, q)

	}
	return &GeoDocumentIterator{
		itr:      iterators,
		Location: location,
		Radius:   radius,
	}
}

type GeoDocumentIterator struct {
	Location *latlng.LatLng
	Radius   uint
	itr      []*firestore.DocumentIterator
}

func (g *GeoDocumentIterator) Next() (*GeoSearchResult, error) {
	if len(g.itr) < 1 {
		log.Printf("All done.")
		return nil, iterator.Done
	}

	di := g.itr[0]
	ds, err := di.Next()

	if err != nil && err != iterator.Done {
		g.Stop()
		return nil, err
	}

	if err == iterator.Done {
		di.Stop()
		g.itr = append(g.itr[:0], g.itr[1:]...)
		// finally recurse.
		return g.Next()
	}

	// check if we're in the right place.
	gd := &GeoSearchResult{}
	ds.DataTo(gd)
	gd.Doc = ds
	dist := uint(geoutils.CalculateDistance(gd.Location.Latitude, gd.Location.Longitude, g.Location.Latitude, g.Location.Longitude))
	gd.Distance = dist
	if dist > g.Radius {

		return g.Next()
	}

	return gd, nil
}

func (g *GeoDocumentIterator) Stop() {
	for _, di := range g.itr {
		di.Stop()
	}
	g.itr = make([]*firestore.DocumentIterator, 0)
}
