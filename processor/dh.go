package processor

// diffieHellmanPowMod does this math:
// g**x mod n
func diffieHellmanPowMod(g, x, p int) int {
	var r int
	var y int = 1

	for x > 0 {
		r = x % 2
		// Fast exponention
		if r == 1 {
			y = (y * g) % p
		}
		g = g * g % p
		x = x / 2
	}

	return y
}
