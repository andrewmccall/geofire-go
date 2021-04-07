package geofire

import (
	"log"
	"sort"
	"testing"

	"github.com/andrewmccall/geoutils"
	"google.golang.org/genproto/googleapis/type/latlng"
)

func createSearchResult(start *latlng.LatLng, loc *latlng.LatLng) GeoSearchResult {
	res := GeoSearchResult{}
	res.SetLocation(loc)
	res.Distance = uint(geoutils.CalculateDistance(start.Latitude, start.Longitude, loc.Latitude, loc.Longitude))
	log.Printf("Created new GeoSearchResults %+v", res)
	return res
}

func TestSorting(t *testing.T) {

	// let's set a centre point for our 'search'
	centre := &latlng.LatLng{
		Latitude:  53.75734328397976,
		Longitude: -2.0180395172300716,
	}

	// lets make a few search results.
	results := make([]GeoSearchResult, 4)
	results[0] = createSearchResult(centre, &latlng.LatLng{
		Latitude:  53.579461,
		Longitude: -1.795476,
	})
	results[1] = createSearchResult(centre, &latlng.LatLng{
		Latitude:  53.312827,
		Longitude: -1.261199,
	})
	results[2] = createSearchResult(centre, &latlng.LatLng{
		Latitude:  53.579461,
		Longitude: -1.86532,
	})
	results[3] = createSearchResult(centre, &latlng.LatLng{
		Latitude:  53.533778,
		Longitude: -1.428852,
	})

	byDistance := ByDistance(results)

	// test the len method is sane
	if len(results) != len(byDistance) {
		t.Error("The len() calls are inconsistent")
	}

	// test the swap method works
	if byDistance[0].GeoHash != "gcw8z1udfr" || byDistance[3].GeoHash != "gcwbx3x8c7" {
		t.Error("GeoHashes not as expected.")
	}
	byDistance.Swap(0, 3)

	if byDistance[0].GeoHash != "gcwbx3x8c7" || byDistance[3].GeoHash != "gcw8z1udfr" {
		t.Error("GeoHashes not as expected after swap.")
	}

	// test the less method works
	if !byDistance.Less(0, 1) {
		t.Error("first should be less than second")
	}
	if byDistance.Less(1, 2) {
		t.Error("second should be less than third")
	}
	// sort a list and let's see how that works.
	sort.Sort(byDistance)

	if byDistance[0].GeoHash != "gcw8v9cdyz" || byDistance[1].GeoHash != "gcw8z1udfr" || byDistance[2].GeoHash != "gcwbx3x8c7" || byDistance[3].GeoHash != "gcrp7339ez" {
		t.Error("Sort order is incorrect.")
	}
}
