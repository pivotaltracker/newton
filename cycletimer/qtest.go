package main

func QTest(xs []float64, confidence float64) []float64 {
	if len(xs) < 3 {
		return xs
	}

	min := MinFloat64(xs...)
	max := MaxFloat64(xs...)
	maxUnder := MaxUnder(max, xs...)
	gap := max - maxUnder
	r := max - min

	if gap/r <= QCrit(len(xs), confidence) {
		return xs
	}

	ns := make([]float64, 0, len(xs)-1)
	found := false
	for _, x := range xs {
		if !found && x == max {
			found = true
			continue
		}
		ns = append(ns, x)
	}
	return ns
}

func QCrit(n int, confidence float64) float64 {
	if n > 10 {
		n = 10
	}

	c := int(confidence * 100.0)

	switch {
	case c < 95:
		c = 90
	case c < 99:
		c = 95
	default:
		c = 99
	}

	return qTable[c][n]
}

var qTable = map[int]map[int]float64{
	90: map[int]float64{
		3:  0.941,
		4:  0.765,
		5:  0.642,
		6:  0.560,
		7:  0.507,
		8:  0.468,
		9:  0.437,
		10: 0.412,
	},
	95: map[int]float64{
		3:  0.970,
		4:  0.829,
		5:  0.710,
		6:  0.625,
		7:  0.568,
		8:  0.526,
		9:  0.493,
		10: 0.466,
	},
	99: map[int]float64{
		3:  0.994,
		4:  0.926,
		5:  0.821,
		6:  0.740,
		7:  0.680,
		8:  0.634,
		9:  0.598,
		10: 0.568,
	},
}
