package auto

import (
	"fmt"

	"github.com/mattn/gopher"
)

func init() {
	for i := 1; i <= 10; i++ {
		gopher.Create(fmt.Sprintf("ʕ◔ϖ◔ʔ .oO( I'm No%d )", i))
	}
}
