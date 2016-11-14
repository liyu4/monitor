package format

import (
	"github.com/kevinchen/stringx"
)

func Translate(source float64) string {

	if source >= 1024 {
		source = source / 1024
	} else {
		return B(source)
	}

	if source >= 1024 {
		source = source / 1024
	} else {
		return k(source)
	}

	if source >= 1024 {
		source = source / 1024
	} else {
		return M(source)
	}

	if source >= 1024 {
		source = source / 1024
	} else {
		return G(source)
	}

	return T(source)
}

//disk
func TranslateDir(source float64) string {
	if source >= 1024 {
		source = source / 1024
	} else {
		return k(source)
	}

	if source >= 1024 {
		source = source / 1024
	} else {
		return M(source)
	}

	if source >= 1024 {
		source = source / 1024
	} else {
		return G(source)
	}

	return T(source)

}

func Float64(source float64) string {
	return stringx.MustString(source)
}

func B(source interface{}) string {
	return stringx.MustString(source) + "B"
}

func k(source interface{}) string {
	return stringx.MustString(source) + "k"
}

func M(source interface{}) string {
	return stringx.MustString(source) + "M"
}

func G(source interface{}) string {
	return stringx.MustString(source) + "G"
}

func T(source interface{}) string {
	return stringx.MustString(source) + "T"
}
