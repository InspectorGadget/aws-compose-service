package controllers

import (
	"context"
	"strings"

	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
)

// ParseDownCommand routes the "down" call to the proper service implementation.
func ParseDownCommand(ctx context.Context, opt structs.Options) error {
	service := strings.ToLower(helpers.WithFallbackValue(opt.Service, "rds"))

	switch service {
	case "rds":
		return RDSDown(ctx, opt)
	case "s3":
		return S3Down(ctx, opt)
	default:
		helpers.Error("unsupported service: %s (expected: rds or s3)", service)
		return nil
	}
}
