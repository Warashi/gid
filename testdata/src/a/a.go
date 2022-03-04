package a

import ( // want `not gid'ed`
	"a/local-imports"
	"errors"
	"fmt"

	_ "github.com/Warashi/wrapmsg"
)

func init() {
	fmt.Println(errors.New(local.Dummy))
}
