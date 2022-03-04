package a

import (
	"fmt"
	"errors"
	_ "github.com/Warashi/wrapmsg"
	"a/local-imports"
)

func init() {
	fmt.Println(errors.New(local.Dummy))
}
