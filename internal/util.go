package internal

func NotEmpty[T ~string | ~[]string | ~[][]string](x, y T) (T, bool) {
	if len(x) > 0 {
		return x, true
	}
	return y, len(y) > 0
}
