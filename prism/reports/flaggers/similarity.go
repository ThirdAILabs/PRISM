package flaggers

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
