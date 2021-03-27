package geofire

import (
	"log"
	"testing"
)

func createQuery(start string, end string) *GeoHashQuery {
	return &GeoHashQuery{
		StartValue: start,
		EndValue:   end,
	}
}

func assertTrue(test bool, t *testing.T) {
	if !test {
		t.Errorf("Expected true, was false")
	}
}

func assertFalse(test bool, t *testing.T) {
	if test {
		t.Errorf("Expected true, was false")
	}
}

func assertEquals(expected *GeoHashQuery, query *GeoHashQuery, t *testing.T) {
	if *query != *expected {
		t.Errorf("Expected %v, was %v", query, expected)
	}
}

func TestqueryForGeoHash(t *testing.T) {
	assertEquals(createQuery("60", "6h"), queryForGeoHash("64m9yn96mx", 6), t)
	assertEquals(createQuery("0", "h"), queryForGeoHash("64m9yn96mx", 1), t)
	assertEquals(createQuery("64", "65"), queryForGeoHash("64m9yn96mx", 10), t)
	assertEquals(createQuery("640", "64h"), queryForGeoHash("6409yn96mx", 11), t)
	assertEquals(createQuery("64h", "64~"), queryForGeoHash("64m9yn96mx", 11), t)
	assertEquals(createQuery("6", "6~"), queryForGeoHash("6", 10), t)
	assertEquals(createQuery("64s", "64~"), queryForGeoHash("64z178", 12), t)
	assertEquals(createQuery("64z", "64~"), queryForGeoHash("64z178", 15), t)
}

func TestCanJoinWith(t *testing.T) {
	assertTrue(createQuery("abcd", "abce").canJoinWith(createQuery("abce", "abcf")), t)
	assertTrue(createQuery("abce", "abcf").canJoinWith(createQuery("abcd", "abce")), t)
	assertTrue(createQuery("abcd", "abcf").canJoinWith(createQuery("abcd", "abce")), t)
	assertTrue(createQuery("abcd", "abcf").canJoinWith(createQuery("abce", "abcf")), t)
	assertTrue(createQuery("abc", "abd").canJoinWith(createQuery("abce", "abcf")), t)
	assertTrue(createQuery("abce", "abcf").canJoinWith(createQuery("abc", "abd")), t)
	assertTrue(createQuery("abcd", "abce~").canJoinWith(createQuery("abc", "abd")), t)
	assertTrue(createQuery("abcd", "abce~").canJoinWith(createQuery("abce", "abcf")), t)
	assertTrue(createQuery("abcd", "abcf").canJoinWith(createQuery("abce", "abcg")), t)

	assertFalse(createQuery("abcd", "abce").canJoinWith(createQuery("abcg", "abch")), t)
	assertFalse(createQuery("abcd", "abce").canJoinWith(createQuery("dce", "dcf")), t)
	assertFalse(createQuery("abc", "abd").canJoinWith(createQuery("dce", "dcf")), t)
}

func createJoinedQueryForTest(s1 string, s2 string, s3 string, s4 string, t *testing.T) *GeoHashQuery {
	q, err := createQuery(s1, s2).joinWith(createQuery(s3, s4))
	if err != nil {
		log.Printf("Error joining queries %s", err.Error())
		t.Fail()
	}
	return q
}

func TestJoinWith(t *testing.T) {

	createQuery("abcd", "abce").joinWith(createQuery("abce", "abcf"))

	assertEquals(createQuery("abcd", "abcf"), createJoinedQueryForTest("abcd", "abce", "abce", "abcf", t), t)
	assertEquals(createQuery("abcd", "abcf"), createJoinedQueryForTest("abce", "abcf", "abcd", "abce", t), t)
	assertEquals(createQuery("abcd", "abcf"), createJoinedQueryForTest("abcd", "abcf", "abcd", "abce", t), t)
	assertEquals(createQuery("abcd", "abcf"), createJoinedQueryForTest("abcd", "abcf", "abce", "abcf", t), t)
	assertEquals(createQuery("abc", "abd"), createJoinedQueryForTest("abc", "abd", "abce", "abcf", t), t)
	assertEquals(createQuery("abc", "abd"), createJoinedQueryForTest("abce", "abcf", "abc", "abd", t), t)
	assertEquals(createQuery("abc", "abd"), createJoinedQueryForTest("abcd", "abce~", "abc", "abd", t), t)
	assertEquals(createQuery("abcd", "abcf"), createJoinedQueryForTest("abcd", "abce~", "abce", "abcf", t), t)
	assertEquals(createQuery("abcd", "abcg"), createJoinedQueryForTest("abcd", "abcf", "abce", "abcg", t), t)

	_, err := createQuery("abcd", "abce").joinWith(createQuery("abcg", "abch"))
	if err != nil {
		log.Printf("Error not returned")
		t.Fail()
	}

	_, err2 := createQuery("abcd", "abce").joinWith(createQuery("dce", "dcf"))
	if err2 != nil {
		log.Printf("Error not returned")
		t.Fail()
	}

	_, err3 := createQuery("abc", "abd").joinWith(createQuery("dce", "dcf"))
	if err3 != nil {
		log.Printf("Error not returned")
		t.Fail()
	}
}
