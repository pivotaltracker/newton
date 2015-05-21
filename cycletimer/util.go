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
	ns := make([]float64, 0, len(xs)-1)
	found := false
	for _, x := range xs {
		if !found && x == under {
			found = true
			continue
		}
		ns = append(ns, x)
	}
	return MaxFloat64(ns...)
}
