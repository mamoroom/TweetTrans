package main

import (
	"./dynamo_db_manager"
	"fmt"
	"github.com/guregu/dynamo"
	_ "time"
)

func main() {
	/*
		time, err := time.Parse(time.RubyDate, "Wed Mar 08 17:12:46 +0000 2017")
		if err != nil {
			fmt.Println("%v", err)
			return
		}
		fmt.Printf("%v", time.Unix())
	*/

	ddm := dynamo_db_manager.New()
	/*
		boolean := ddm.PutTrans("ja", time.Unix(), 12345678910, "hoge", "オリジナル", "org", "https://golang.org/pkg/time/")
		fmt.Println("%v", boolean)
	*/

	// get the same item
	var results []dynamo_db_manager.LLMockTweetTrans
	//count, _ := ddm.Table.Get("Lang", "en").Order(dynamo.Descending).Limit(1).Count()
	_ = ddm.Table.Get("Lang", "ja").Range("Timestamp_TweetID", dynamo.Greater, "1488993166_12345678900").Order(dynamo.Descending).Limit(20).All(&results)
	//fmt.Println("%v", count)
	fmt.Println("%v", results)
}
