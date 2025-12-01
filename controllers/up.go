package controllers

import (
	"context"
	"strings"

	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
)

// ParseUpCommand routes the "up" call to the proper service implementation.
func ParseUpCommand(ctx context.Context, opt structs.Options) error {
	if opt.Service == "" {
		helpers.Error("service type must be specified")
	}
	if opt.Region == "" {
		helpers.Error("region must be specified")
	}

	service := strings.ToLower(opt.Service)

	switch service {
	case "rds":
		return RDSUp(ctx, opt)
	case "s3":
		return S3Up(ctx, opt)
	default:
		helpers.Error("unsupported service: %s (expected: rds or s3)", service)
		return nil
	}
}
