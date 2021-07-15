package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func (h *handler) create(ctx context.Context, detail CloudTrailDetail) error {
	resp := CreateNetworkInterfaceResponseElements{}
	err := json.Unmarshal(detail.ResponseElements, &resp)
	if err != nil {
		return errors.WithStack(err)
	}

	eni := resp.NetworkInterface.NetworkInterfaceId

	allocate, err := h.api.AllocateAddressWithContext(ctx, &ec2.AllocateAddressInput{
		Domain: aws.String(ec2.DomainTypeVpc),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeElasticIp),
				Tags: []*ec2.Tag{
					{Key: aws.String("lambdaeip:owned"), Value: aws.String("true")},
					{Key: aws.String("lambdaeip:eni"), Value: aws.String(eni)},
				},
			},
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	associate, err := h.api.AssociateAddressWithContext(ctx, &ec2.AssociateAddressInput{
		AllocationId:       allocate.AllocationId,
		NetworkInterfaceId: &eni,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	j, _ := json.Marshal(map[string]string{
		"eni":         eni,
		"eip":         *allocate.AllocationId,
		"association": *associate.AssociationId,
	})
	fmt.Println(string(j))

	return nil
}

type CreateNetworkInterfaceResponseElements struct {
	NetworkInterface struct {
		NetworkInterfaceId string `json:"networkInterfaceId"`
	} `json:"networkInterface"`
}
