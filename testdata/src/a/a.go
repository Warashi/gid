package a

import ( // want `not gid'ed`
	"fmt"
	"errors"

	_ "github.com/Warashi/wrapmsg"

	"a/local-imports"
)

func init() {
	fmt.Println(errors.New(local.Dummy))
}
