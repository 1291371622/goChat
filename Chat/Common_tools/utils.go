package Common_tools

import (
	"fmt"
	"os"
)

func HanddleError(err error,when string){
	if err!=nil {
		fmt.Println(when,err)
		os.Exit(1)
	}
}
