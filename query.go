package geofire

import (
	"math"
	"strings"

	"github.com/andrewmccall/geoutils"
	"google.golang.org/genproto/googleapis/type/latlng"
)

// Length of a degree latitude at the equator
const METERS_PER_DEGREE_LATITUDE = float64(110574)

// The equatorial circumference of the earth in meters
const EARTH_MERIDIONAL_CIRCUMFERENCE = float64(40007860)

// The equatorial radius of the earth in meters
const EARTH_EQ_RADIUS = float64(6378137)

// The meridional radius of the earth in meters
const EARTH_POLAR_RADIUS = float64(6357852.3)

/* The following value assumes a polar radius of
* r_p = 6356752.3
* and an equatorial radius of
* r_e = 6378137
* The value is calculated as e2 == (r_e^2 - r_p^2)/(r_e^2)
* Use exact value to avoid rounding errors
 */
const EARTH_E2 = float64(0.00669447819799)

const EPSILON = 1e-12

const DEFAULT_PRECISION = 10

const MAX_PRECISION_BITS = 22

type GeoHashQuery struct {
	StartValue string
	EndValue   string
}

func (g *GeoHashQuery) isPrefix(other *GeoHashQuery) bool {
	return (strings.Compare(other.EndValue, g.StartValue) >= 0) &&
		(strings.Compare(other.StartValue, g.StartValue) < 0) &&
		(strings.Compare(other.EndValue, g.EndValue) < 0)
}

func (g *GeoHashQuery) isSuperQuery(other *GeoHashQuery) bool {
	startCompare := strings.Compare(other.StartValue, g.StartValue)
	return startCompare <= 0 && strings.Compare(other.EndValue, g.EndValue) >= 0
}

func (g *GeoHashQuery) canJoinWith(other *GeoHashQuery) bool {
	return g.isPrefix(other) || other.isPrefix(g) || g.isSuperQuery(other) || other.isSuperQuery(g)
}

func (g *GeoHashQuery) joinWith(other *GeoHashQuery) (*GeoHashQuery, error) {
	if other.isPrefix(g) {
		return &GeoHashQuery{
			StartValue: g.StartValue,
			EndValue:   other.EndValue,
		}, nil
	} else if g.isPrefix(other) {
		return &GeoHashQuery{
			StartValue: other.StartValue,
			EndValue:   g.EndValue,
		}, nil
	} else if g.isSuperQuery(other) {
		return other, nil
	} else if other.isSuperQuery(g) {
		return g, nil
	} else {
		return nil, nil
	}
}

func queryForGeoHash(geoHash string, bits float64) *GeoHashQuery {
	precision := int(math.Ceil(bits / BITS_PER_BASE32_CHAR))
	if len(geoHash) < precision {
		return &GeoHashQuery{
			StartValue: geoHash,
			EndValue:   geoHash + "~",
		}
	}
	hash := geoHash[0:precision]
	base := hash[0 : len(hash)-1]
	lastValue := ToBase32Value(byte(hash[len(hash)-1]))
	significantBits := bits - float64(len(base)*BITS_PER_BASE32_CHAR)
	unusedBits := int(BITS_PER_BASE32_CHAR - significantBits)
	// delete unused bits
	StartValue := (lastValue >> unusedBits) << unusedBits
	EndValue := StartValue + (1 << unusedBits)
	startHash := base + string(ToBase32Char(StartValue))
	var endHash string
	if EndValue > 31 {
		endHash = base + string('~')
	} else {
		endHash = base + string(ToBase32Char(EndValue))
	}
	return &GeoHashQuery{
		StartValue: startHash,
		EndValue:   endHash,
	}
}

//GetQueryBounds returns the start and end values for the geohash bounds with a given radius.
func QueryiesAtLocation(location *latlng.LatLng, radius float64) map[GeoHashQuery]struct{} {
	queryBits := math.Max(1, bitsForBoundingBox(location, radius))
	geoHashPrecision := uint(math.Ceil(queryBits / BITS_PER_BASE32_CHAR))

	latitude := location.Latitude
	longitude := location.Longitude
	latitudeDegrees := radius / METERS_PER_DEGREE_LATITUDE
	latitudeNorth := math.Min(90, latitude+latitudeDegrees)
	latitudeSouth := math.Max(-90, latitude-latitudeDegrees)
	longitudeDeltaNorth := distanceToLongitudeDegrees(radius, latitudeNorth)
	longitudeDeltaSouth := distanceToLongitudeDegrees(radius, latitudeSouth)
	longitudeDelta := math.Max(longitudeDeltaNorth, longitudeDeltaSouth)

	queries := make(map[GeoHashQuery]struct{})

	geoHash, _ := geoutils.GeoHash(latitude, longitude, geoHashPrecision)
	geoHashW, _ := geoutils.GeoHash(latitude, wrapLongitude(longitude-longitudeDelta), geoHashPrecision)
	geoHashE, _ := geoutils.GeoHash(latitude, wrapLongitude(longitude+longitudeDelta), geoHashPrecision)

	geoHashN, _ := geoutils.GeoHash(latitudeNorth, longitude, geoHashPrecision)
	geoHashNW, _ := geoutils.GeoHash(latitudeNorth, wrapLongitude(longitude-longitudeDelta), geoHashPrecision)
	geoHashNE, _ := geoutils.GeoHash(latitudeNorth, wrapLongitude(longitude+longitudeDelta), geoHashPrecision)

	geoHashS, _ := geoutils.GeoHash(latitudeSouth, longitude, geoHashPrecision)
	geoHashSW, _ := geoutils.GeoHash(latitudeSouth, wrapLongitude(longitude-longitudeDelta), geoHashPrecision)
	geoHashSE, _ := geoutils.GeoHash(latitudeSouth, wrapLongitude(longitude+longitudeDelta), geoHashPrecision)

	queries[*queryForGeoHash(geoHash, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashE, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashW, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashN, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashNE, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashNW, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashS, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashSE, queryBits)] = struct{}{}
	queries[*queryForGeoHash(geoHashSW, queryBits)] = struct{}{}

	// Join queries
	for didJoin := true; didJoin; {
		var query1 *GeoHashQuery = nil
		var query2 *GeoHashQuery = nil
		for query, _ := range queries {
			for other, _ := range queries {
				if query != other && query.canJoinWith(&other) {
					query1 = &query
					query2 = &other
					break
				}
			}
		}
		if query1 != nil && query2 != nil {

			delete(queries, *query1)
			delete(queries, *query2)
			joined, _ := query1.joinWith(query2)
			queries[*joined] = struct{}{}
			didJoin = true
		} else {
			didJoin = false
		}
	}

	return queries
}

func wrapLongitude(longitude float64) float64 {
	if longitude >= -180 && longitude <= 180 {
		return longitude
	}
	adjusted := longitude + 180
	if adjusted > 0 {
		return math.Mod(adjusted, 360.0) - 180
	} else {
		return 180 - math.Mod((adjusted*-1), 360)
	}
}

func bitsLatitude(resolution float64) float64 {

	return math.Min(math.Log(EARTH_MERIDIONAL_CIRCUMFERENCE/2/resolution)/math.Log(2), MAX_PRECISION_BITS)
}

func bitsLongitude(resolution float64, latitude float64) float64 {
	degrees := distanceToLongitudeDegrees(resolution, latitude)
	if math.Abs(degrees) > 0 {
		return math.Max(1, math.Log(360/degrees)/math.Log(2))
	}
	return 1
}

func bitsForBoundingBox(location *latlng.LatLng, size float64) float64 {
	latitudeDegreesDelta := distanceToLatitudeDegrees(size)
	latitudeNorth := math.Min(90, location.Latitude+latitudeDegreesDelta)
	latitudeSouth := math.Max(-90, location.Latitude-latitudeDegreesDelta)
	bitsLatitude := math.Floor(bitsLatitude(size)) * 2
	bitsLongitudeNorth := math.Floor(bitsLongitude(size, latitudeNorth))*2 - 1
	bitsLongitudeSouth := math.Floor(bitsLongitude(size, latitudeSouth))*2 - 1
	return math.Min(bitsLatitude, math.Min(bitsLongitudeNorth, bitsLongitudeSouth))
}

func distanceToLatitudeDegrees(distance float64) float64 {
	return distance / METERS_PER_DEGREE_LATITUDE
}

func distanceToLongitudeDegrees(distance float64, latitude float64) float64 {
	radians := toRadians(latitude)
	numerator := math.Cos(radians) * EARTH_EQ_RADIUS * math.Pi / 180
	denominator := 1 / math.Sqrt(1-EARTH_E2*math.Sin(radians)*math.Sin(radians))
	deltaDegrees := numerator * denominator
	if deltaDegrees < EPSILON {
		if distance > 0 {
			return 360
		} else {
			return distance
		}
	} else {
		return math.Min(360, distance/deltaDegrees)
	}
}

func toRadians(deg float64) float64 {
	return float64(deg) * (math.Pi / 180.0)
}
