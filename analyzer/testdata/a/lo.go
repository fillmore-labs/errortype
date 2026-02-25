package a

import "github.com/samber/lo"

func LoErrors(err error) {
	_, _ = lo.ErrorsAs[*ValueError](err) // want ` \(et:ast\)$`
}
