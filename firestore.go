package geofire

import (
	"context"

	firestore "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/genproto/googleapis/type/latlng"
)

type GeoDocument struct {
	geoHash  string         `firestore:"g,omitempty"`
	Location *latlng.LatLng `firestore:"l,omitempty"`
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
		itr: iterators,
	}
}

type GeoDocumentIterator struct {
	itr []*firestore.DocumentIterator
}

func (g *GeoDocumentIterator) Next() (*firestore.DocumentSnapshot, error) {
	if len(g.itr) < 1 {
		return nil, iterator.Done
	}
	// needs to iterate across each iterator, if one runs out move to the next.
	// each document needs to be checked to be sure it's in the right range.
	// return it.
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
	// if we empty the whole thing, return iterator.Done
	return ds, nil
}

func (g *GeoDocumentIterator) Stop() {
	for _, di := range g.itr {
		di.Stop()
	}
	g.itr = make([]*firestore.DocumentIterator, 0)
}
