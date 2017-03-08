package main

import (
	"./dynamo_db_manager"
	"fmt"
	"time"
)

func main() {
	time, err := time.Parse(time.RubyDate, "Wed Mar 08 17:12:46 +0000 2017")
	if err != nil {
		fmt.Println("%v", err)
		return
	}

	ddm := dynamo_db_manager.New()
	boolean := ddm.PutTrans("ja", time.Unix(), 12345678910, "hoge", "オリジナル", "org", "https://golang.org/pkg/time/")
	fmt.Println("%v", boolean)
}
