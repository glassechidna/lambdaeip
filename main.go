package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/pkg/errors"
	"os"
)

func main() {
	sess, err := session.NewSession()
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	h := &handler{
		sentinelGroupId: os.Getenv("SENTINEL_SECURITY_GROUP_ID"),
		api:             ec2.New(sess),
	}

	lambda.Start(h.handle)
}

type handler struct {
	sentinelGroupId string
	api             ec2iface.EC2API
}

func (h *handler) handle(ctx context.Context, event *events.CloudWatchEvent) error {
	fmt.Println(string(event.Detail))

	detail := CloudTrailDetail{}
	err := json.Unmarshal(event.Detail, &detail)
	if err != nil {
		return errors.WithStack(err)
	}

	switch detail.EventName {
	case "CreateNetworkInterface":
		return h.create(ctx, detail)
	case "DeleteNetworkInterface":
		return h.delete(ctx, detail)
	}

	return nil
}

type CloudTrailDetail struct {
	EventName         string          `json:"eventName"`
	RequestParameters json.RawMessage `json:"requestParameters"`
	ResponseElements  json.RawMessage `json:"responseElements"`
}
