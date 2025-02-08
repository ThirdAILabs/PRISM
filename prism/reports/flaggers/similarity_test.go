package flaggers

import (
	"math"
	"testing"
)

func TestLevensteinDistance(t *testing.T) {
	tests := []struct {
		s1           string
		s2           string
		expectedDist int
	}{
		{"", "hello", 5},
		{"hello", "", 5},
		{"hello", "hello", 0},
		{"ab", "aa", 1},
		{"ab", "ba", 2},
		{"ab", "aaa", 2},
		{"bbb", "a", 3},
		{"kitten", "sitting", 3},
		{"distance", "difference", 5},
		{"levenshtein", "frankenstein", 6},
		{"resume and cafe", "resumes and cafes", 2},
		{"a very long string that is meant to exceed", "another very long string that is meant to exceed", 6},
	}

	for _, test := range tests {
		dist := levenshteinDistance(test.s1, test.s2, 1)
		if dist != test.expectedDist {
			t.Errorf("s1=%s s2=%s expected=%d actual=%d", test.s1, test.s2, test.expectedDist, dist)
		}
	}

	if levenshteinDistance("abc", "axc", 1) != 1 {
		t.Errorf("incorrect dist with weight 1")
	}

	if levenshteinDistance("abc", "axc", 2) != 2 {
		t.Errorf("incorrect dist with weight 2")
	}
}

func TestIndelSim(t *testing.T) {
	tests := []struct {
		s1          string
		s2          string
		expectedSim float64
	}{
		{"", "hello", 0},
		{"hello", "", 0},
		{"hello", "hello", 1},
		{"ab", "aa", 0.5},
		{"aa", "ab", 0.5},
		{"abc", "ab", 0.8},
		{"ab", "abc", 0.8},
	}

	for _, test := range tests {
		sim := IndelSimilarity(test.s1, test.s2)
		if sim != test.expectedSim {
			t.Errorf("s1=%s s2=%s expected=%f actual=%f", test.s1, test.s2, test.expectedSim, sim)
		}
	}
}

func TestJaroSim(t *testing.T) {
	tests := []struct {
		s1          string
		s2          string
		expectedSim float64
	}{
		{s1: "SHACKLEFORD", s2: "SHACKELFORD", expectedSim: 0.98182},
		{s1: "DUNNINGHAM", s2: "CUNNIGHAM", expectedSim: 0.89630},
		{s1: "NICHLESON", s2: "NICHULSON", expectedSim: 0.95556},
		{s1: "JONES", s2: "JOHNSON", expectedSim: 0.83238},
		{s1: "MASSEY", s2: "MASSIE", expectedSim: 0.93333},
		{s1: "ABROMS", s2: "ABRAMS", expectedSim: 0.92222},
		{s1: "HARDIN", s2: "MARTINEZ", expectedSim: 0.72222},
		{s1: "ITMAN", s2: "SMITH", expectedSim: 0.46667},
		{s1: "JERALDINE", s2: "GERALDINE", expectedSim: 0.92593},
		{s1: "MARTHA", s2: "MARHTA", expectedSim: 0.96111},
		{s1: "MICHELLE", s2: "MICHAEL", expectedSim: 0.92143},
		{s1: "JULIES", s2: "JULIUS", expectedSim: 0.93333},
		{s1: "TANYA", s2: "TONYA", expectedSim: 0.88000},
		{s1: "DWAYNE", s2: "DUANE", expectedSim: 0.84000},
		{s1: "SEAN", s2: "SUSAN", expectedSim: 0.80500},
		{s1: "JON", s2: "JOHN", expectedSim: 0.93333},
		{s1: "JON", s2: "JAN", expectedSim: 0.80000},
		{s1: "DWAYNE", s2: "DYUANE", expectedSim: 0.84000},
		{s1: "CRATE", s2: "TRACE", expectedSim: 0.73333},
		{s1: "WIBBELLY", s2: "WOBRELBLY", expectedSim: 0.85298},
		{s1: "DIXON", s2: "DICKSONX", expectedSim: 0.81333},
		{s1: "MARHTA", s2: "MARTHA", expectedSim: 0.96111},
		{s1: "AL", s2: "AL", expectedSim: 1.00000},
		{s1: "aaaaaabc", s2: "aaaaaabd", expectedSim: 0.95000},
		{s1: "ABCVWXYZ", s2: "CABVWXYZ", expectedSim: 0.95833},
		{s1: "ABCAWXYZ", s2: "BCAWXYZ", expectedSim: 0.91071},
		{s1: "ABCVWXYZ", s2: "CBAWXYZ", expectedSim: 0.91071},
		{s1: "ABCDUVWXYZ", s2: "DABCUVWXYZ", expectedSim: 0.93333},
		{s1: "ABCDUVWXYZ", s2: "DBCAUVWXYZ", expectedSim: 0.96667},
		{s1: "ABBBUVWXYZ", s2: "BBBAUVWXYZ", expectedSim: 0.96667},
		{s1: "ABCDUV11lLZ", s2: "DBCAUVWXYZ", expectedSim: 0.73117},
		{s1: "ABBBUVWXYZ", s2: "BBB11L3VWXZ", expectedSim: 0.77879},
		{s1: "A", s2: "A", expectedSim: 1.00000},
		{s1: "AB", s2: "AB", expectedSim: 1.00000},
		{s1: "ABC", s2: "ABC", expectedSim: 1.00000},
		{s1: "ABCD", s2: "ABCD", expectedSim: 1.00000},
		{s1: "ABCDE", s2: "ABCDE", expectedSim: 1.00000},
		{s1: "AA", s2: "AA", expectedSim: 1.00000},
		{s1: "AAA", s2: "AAA", expectedSim: 1.00000},
		{s1: "AAAA", s2: "AAAA", expectedSim: 1.00000},
		{s1: "AAAAA", s2: "AAAAA", expectedSim: 1.00000},
		{s1: "A", s2: "B", expectedSim: 0.00000},
		{s1: "", s2: "ABC", expectedSim: 0.00000},
		{s1: "ABCD", s2: "", expectedSim: 0.00000},
		{s1: "", s2: "", expectedSim: 0.00000},
	}

	for _, test := range tests {
		sim := JaroWinklerSimilarity(test.s1, test.s2)
		if math.Round(1000*sim) != math.Round(1000*test.expectedSim) {
			t.Errorf("s1=%s s2=%s expected=%f actual=%f", test.s1, test.s2, test.expectedSim, sim)
		}
	}
}
