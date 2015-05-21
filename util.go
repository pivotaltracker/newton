package main

func MinFloat64(xs ...float64) float64 {
	m := xs[0]
	for _, x := range xs {
		if x < m {
			m = x
		}
	}
	return m
}

func MaxFloat64(xs ...float64) float64 {
	m := xs[0]
	for _, x := range xs {
		if x > m {
			m = x
		}
	}
	return m
}

func MaxUnder(under float64, xs ...float64) float64 {
	m := xs[0]
	if m >= under {
		for _, x := range xs {
			if x < under {
				m = x
				break
			}
		}
	}
	for _, x := range xs {
		if x > m && x < under {
			m = x
		}
	}
	return m
}
