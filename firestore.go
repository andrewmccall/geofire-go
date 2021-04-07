package geofire

import (
	"context"
	"log"
	"sort"

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

type GeoQuery struct {
	baseQuery *firestore.Query
	location  *latlng.LatLng
	radius    uint
}

func NewGeoQuery(baseQuery *firestore.Query, location *latlng.LatLng, radius uint) *GeoQuery {
	return &GeoQuery{
		baseQuery: baseQuery,
		location:  location,
		radius:    radius,
	}

}

func (q *GeoQuery) Documents(ctx context.Context) *GeoDocumentIterator {

	queries := QueryiesAtLocation(q.location, float64(q.radius))

	log.Printf("Running with base query %v", q.baseQuery)

	var iterators []*firestore.DocumentIterator
	for query := range queries {
		exec := q.baseQuery.OrderBy("g", firestore.Asc).StartAt(query.StartValue).EndAt(query.EndValue)
		iterators = append(iterators, exec.Documents(ctx))
	}
	return &GeoDocumentIterator{
		itr:      iterators,
		Location: q.location,
		Radius:   q.radius,
	}
}

type GeoCollectionRef firestore.CollectionRef

// Closest documents
func (iterator *GeoDocumentIterator) Closest(max int) ByDistance {

	results := ByDistance(make([]GeoSearchResult, 0))
	for {
		c, err := iterator.Next()
		if err != nil {
			log.Printf("ERR %v", err)
			break
		}
		if len(results) < max {
			results = append(results, *c)
			continue
		}

		if c.Distance < results[len(results)-1].Distance {
			results[len(results)-1] = *c
		}

		sort.Sort(results)
	}

	return results
}

// ByDistance type provided for sorting search results
type ByDistance []GeoSearchResult

func (r ByDistance) Len() int           { return len(r) }
func (r ByDistance) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByDistance) Less(i, j int) bool { return r[i].Distance < r[j].Distance }

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
