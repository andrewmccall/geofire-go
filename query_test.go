package geofire

import (
	"testing"
)

func createQuery(start string, end string) *GeoHashQuery {
	return &GeoHashQuery{
		startValue: start,
		endValue:   end,
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
	if &query != &expected {
		t.Errorf("Expected %v, was %v", query, expected)
	}
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

func TestJoinWith(t *testing.T) {

	assertEquals(createQuery("abcd", "abcf"), createQuery("abcd", "abce")).joinWith(createQuery("abce", "abcf")), t)
	assertEquals(createQuery("abcd", "abcf"), createQuery("abce", "abcf").joinWith(createQuery("abcd", "abce")), t)
	assertEquals(createQuery("abcd", "abcf"), createQuery("abcd", "abcf").joinWith(createQuery("abcd", "abce")), t)
	assertEquals(createQuery("abcd", "abcf"), createQuery("abcd", "abcf").joinWith(createQuery("abce", "abcf")), t)
	assertEquals(createQuery("abc", "abd"), createQuery("abc", "abd").joinWith(createQuery("abce", "abcf")), t)
	assertEquals(createQuery("abc", "abd"), createQuery("abce", "abcf").joinWith(createQuery("abc", "abd")), t)
	assertEquals(createQuery("abc", "abd"), createQuery("abcd", "abce~").joinWith(createQuery("abc", "abd")), t)
	assertEquals(createQuery("abcd", "abcf"), createQuery("abcd", "abce~").joinWith(createQuery("abce", "abcf")), t)
	assertEquals(createQuery("abcd", "abcg"), createQuery("abcd", "abcf").joinWith(createQuery("abce", "abcg")), t)

	// try {
	// 	createQuery("abcd", "abce").joinWith(createQuery("abcg", "abch"));
	// 	Assert.fail("Exception was not thrown!");
	// } catch(IllegalArgumentException expected) {
	// }

	// try {
	// 	createQuery("abcd", "abce").joinWith(createQuery("dce", "dcf"));
	// 	Assert.fail("Exception was not thrown!");
	// } catch(IllegalArgumentException expected) {
	// }

	// try {
	// 	createQuery("abc", "abd").joinWith(createQuery("dce", "dcf"));
	// 	Assert.fail("Exception was not thrown!");
	// } catch(IllegalArgumentException expected) {
	// }
}
