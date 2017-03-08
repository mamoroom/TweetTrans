package dynamo_db_manager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"strconv"
)

type DynamoDBManager struct {
	DynamoDB dynamo.DB
	Table    dynamo.Table
}

type LLMockTweetTrans struct {
	Lang              string
	Timestamp_TweetID string
	ScreenName        string
	OrgTweet          string
	TransedTweet      string
	ProfileImageURL   string
}

func New() *DynamoDBManager {
	ddm := new(DynamoDBManager)
	ddm.DynamoDB = *dynamo.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})
	ddm.Table = ddm.DynamoDB.Table("ll_mock_tweet_trans")
	return ddm
}

func (ddm *DynamoDBManager) PutTrans(lang string, unixtime int64, tweet_id int64, screen_name string, org_tweet string, transed_tweet string, profile_image_url string) bool {
	data := LLMockTweetTrans{
		lang,
		strconv.FormatInt(unixtime, 10) + "_" + strconv.FormatInt(tweet_id, 10),
		screen_name,
		org_tweet,
		transed_tweet,
		profile_image_url,
	}

	err := ddm.Table.Put(data).Run()
	if err != nil {
		fmt.Println("Can't put data")
		return false
	}
	return true
}
