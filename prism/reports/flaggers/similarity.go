package flaggers

func levenshteinDistance(s1, s2 string, subWeight int) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	if s1 == s2 {
		return 0
	}

	// len(s1) = min(len(s1), len(s2))
	if len(s2) < len(s1) {
		s1, s2 = s2, s1
	}

	l1, l2 := len(s1), len(s2)

	prevRow := make([]int, l1+1)
	for i := range prevRow {
		prevRow[i] = i
	}

	for j := 1; j <= l2; j++ {
		last := j
		for i := 1; i <= l1; i++ {
			current := prevRow[i-1]
			if s1[i-1] != s2[j-1] {
				current = min(prevRow[i-1]+subWeight, prevRow[i]+1, last+1)
			}
			prevRow[i-1] = last
			last = current
		}
		prevRow[l1] = last
	}

	return prevRow[l1]
}

func IndelSimilarity(s1, s2 string) float64 {
	return 1 - float64(levenshteinDistance(s1, s2, 2))/(float64(len(s1)+len(s2)))
}

func JaroWinklerSimilarity(s1, s2 string) float64 {
	l1, l2 := len(s1), len(s2)

	if l1 == 0 || l2 == 0 {
		return 0
	}

	if s1 == s2 {
		return 1.0
	}

	maxDist := (max(l1, l2) / 2) - 1

	s1Matched, s2Matched := make([]bool, l1), make([]bool, l2)

	m := 0
	for i := 0; i < l1; i++ {
		for j := max(i-maxDist, 0); j < min(l2, i+maxDist+1); j++ {
			if s1[i] == s2[j] && !s2Matched[j] {
				s1Matched[i] = true
				s2Matched[j] = true
				m++
				break
			}
		}
	}

	if m == 0 {
		return 0
	}

	t := 0

	j := 0
	for i := 0; i < l1; i++ {
		if s1Matched[i] {
			for ; !s2Matched[j]; j++ {
			}

			if s1[i] != s2[j] {
				t++
			}
			j++
		}
	}
	t /= 2

	mFloat := float64(m)
	tFloat := float64(t)

	sim := (mFloat/float64(l1) + mFloat/float64(l2) + (mFloat-tFloat)/mFloat) / 3

	limit := min(l1, l2, 4)
	prefixCnt := 0
	for ; prefixCnt < limit && s1[prefixCnt] == s2[prefixCnt]; prefixCnt++ {
	}

	return sim + float64(prefixCnt)*0.1*(1-sim)
}
