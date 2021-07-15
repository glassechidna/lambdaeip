package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func (h *handler) delete(ctx context.Context, detail CloudTrailDetail) error {
	req := DeleteNetworkInterfaceRequestParameters{}
	err := json.Unmarshal(detail.RequestParameters, &req)
	if err != nil {
		return errors.WithStack(err)
	}

	describe, err := h.api.DescribeAddressesWithContext(ctx, &ec2.DescribeAddressesInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("tag:lambdaeip:eni"), Values: aws.StringSlice([]string{req.NetworkInterfaceId})},
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}

	for _, address := range describe.Addresses {
		_, _ = h.api.DisassociateAddressWithContext(ctx, &ec2.DisassociateAddressInput{
			AssociationId: address.AssociationId,
		})

		_, err = h.api.ReleaseAddressWithContext(ctx, &ec2.ReleaseAddressInput{
			AllocationId: address.AllocationId,
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

type DeleteNetworkInterfaceRequestParameters struct {
	NetworkInterfaceId string `json:"networkInterfaceId"`
}
