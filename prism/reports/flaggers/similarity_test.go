package flaggers

import (
	"math"
	"testing"
)

func TestJaroSim(t *testing.T) {
	tests := []struct {
		s1           string
		s2           string
		expected_sim float64
	}{
		{s1: "SHACKLEFORD", s2: "SHACKELFORD", expected_sim: 0.98182},
		{s1: "DUNNINGHAM", s2: "CUNNIGHAM", expected_sim: 0.89630},
		{s1: "NICHLESON", s2: "NICHULSON", expected_sim: 0.95556},
		{s1: "JONES", s2: "JOHNSON", expected_sim: 0.83238},
		{s1: "MASSEY", s2: "MASSIE", expected_sim: 0.93333},
		{s1: "ABROMS", s2: "ABRAMS", expected_sim: 0.92222},
		{s1: "HARDIN", s2: "MARTINEZ", expected_sim: 0.72222},
		{s1: "ITMAN", s2: "SMITH", expected_sim: 0.46667},
		{s1: "JERALDINE", s2: "GERALDINE", expected_sim: 0.92593},
		{s1: "MARTHA", s2: "MARHTA", expected_sim: 0.96111},
		{s1: "MICHELLE", s2: "MICHAEL", expected_sim: 0.92143},
		{s1: "JULIES", s2: "JULIUS", expected_sim: 0.93333},
		{s1: "TANYA", s2: "TONYA", expected_sim: 0.88000},
		{s1: "DWAYNE", s2: "DUANE", expected_sim: 0.84000},
		{s1: "SEAN", s2: "SUSAN", expected_sim: 0.80500},
		{s1: "JON", s2: "JOHN", expected_sim: 0.93333},
		{s1: "JON", s2: "JAN", expected_sim: 0.80000},
		{s1: "DWAYNE", s2: "DYUANE", expected_sim: 0.84000},
		{s1: "CRATE", s2: "TRACE", expected_sim: 0.73333},
		{s1: "WIBBELLY", s2: "WOBRELBLY", expected_sim: 0.85298},
		{s1: "DIXON", s2: "DICKSONX", expected_sim: 0.81333},
		{s1: "MARHTA", s2: "MARTHA", expected_sim: 0.96111},
		{s1: "AL", s2: "AL", expected_sim: 1.00000},
		{s1: "aaaaaabc", s2: "aaaaaabd", expected_sim: 0.95000},
		{s1: "ABCVWXYZ", s2: "CABVWXYZ", expected_sim: 0.95833},
		{s1: "ABCAWXYZ", s2: "BCAWXYZ", expected_sim: 0.91071},
		{s1: "ABCVWXYZ", s2: "CBAWXYZ", expected_sim: 0.91071},
		{s1: "ABCDUVWXYZ", s2: "DABCUVWXYZ", expected_sim: 0.93333},
		{s1: "ABCDUVWXYZ", s2: "DBCAUVWXYZ", expected_sim: 0.96667},
		{s1: "ABBBUVWXYZ", s2: "BBBAUVWXYZ", expected_sim: 0.96667},
		{s1: "ABCDUV11lLZ", s2: "DBCAUVWXYZ", expected_sim: 0.73117},
		{s1: "ABBBUVWXYZ", s2: "BBB11L3VWXZ", expected_sim: 0.77879},
		{s1: "A", s2: "A", expected_sim: 1.00000},
		{s1: "AB", s2: "AB", expected_sim: 1.00000},
		{s1: "ABC", s2: "ABC", expected_sim: 1.00000},
		{s1: "ABCD", s2: "ABCD", expected_sim: 1.00000},
		{s1: "ABCDE", s2: "ABCDE", expected_sim: 1.00000},
		{s1: "AA", s2: "AA", expected_sim: 1.00000},
		{s1: "AAA", s2: "AAA", expected_sim: 1.00000},
		{s1: "AAAA", s2: "AAAA", expected_sim: 1.00000},
		{s1: "AAAAA", s2: "AAAAA", expected_sim: 1.00000},
		{s1: "A", s2: "B", expected_sim: 0.00000},
		{s1: "", s2: "ABC", expected_sim: 0.00000},
		{s1: "ABCD", s2: "", expected_sim: 0.00000},
		{s1: "", s2: "", expected_sim: 0.00000},
	}

	for _, test := range tests {
		sim := JaroWinklerSimilarity(test.s1, test.s2)
		if math.Round(1000*sim) != math.Round(1000*test.expected_sim) {
			t.Errorf("s1=%s s2=%s expected=%f actual=%f", test.s1, test.s2, test.expected_sim, sim)
		}
	}
}
