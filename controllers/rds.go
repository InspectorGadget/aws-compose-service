package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/InspectorGadget/aws-compose-service/helpers"
	"github.com/InspectorGadget/aws-compose-service/structs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

func defaultPortForEngine(engine string) int {
	switch strings.ToLower(engine) {
	case "postgres", "postgresql":
		return 5432
	case "mysql", "mariadb":
		return 3306
	case "sqlserver":
		return 1433
	default:
		return 5432
	}
}

func instanceClassOrDefault(v string) string {
	if v == "" {
		return "db.t3.micro"
	}
	return v
}

func allocatedStorageOrDefault(v int) int32 {
	if v <= 0 {
		return 20
	}
	return int32(v)
}

// RDSUp creates (or reuses) an RDS instance and exports its connection details
// as environment variables for the Compose service.
func RDSUp(ctx context.Context, opt structs.Options) error {
	region := helpers.WithFallbackValue(opt.Region, "ap-southeast-1")
	engine := helpers.WithFallbackValue(opt.Engine, "postgres")
	dbName := helpers.WithFallbackValue(opt.DBName, "app")
	name := helpers.WithFallbackValue(opt.Name, "rds")

	username := helpers.WithFallbackValue(opt.Username, "admin")
	password := helpers.WithFallbackValue(opt.Password, "password")

	cfg, err := helpers.LoadAWSConfig(ctx, region)
	if err != nil {
		helpers.Error("unable to load AWS config: %v", err)
		return err
	}

	client := rds.NewFromConfig(cfg)

	// 1) Check if instance already exists
	describeOut, err := client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(name),
	})
	if err != nil {
		switch err.(type) {
		case *rdstypes.DBInstanceNotFoundFault:
			// not found, proceed to create
		default:
			helpers.Error("describe DB instances failed: %v", err)
		}
	}

	var instance *rdstypes.DBInstance

	if len(describeOut.DBInstances) > 0 {
		instance = &describeOut.DBInstances[0]
		helpers.Info("reusing existing RDS instance %s in %s", name, region)
	} else {
		helpers.Info("creating RDS instance %s in %s (engine=%s)", name, region, engine)

		createInput := &rds.CreateDBInstanceInput{
			DBInstanceIdentifier: aws.String(name),
			Engine:               aws.String(engine),
			MasterUsername:       aws.String(username),
			MasterUserPassword:   aws.String(password),
			DBInstanceClass:      aws.String(instanceClassOrDefault(opt.InstanceClass)),
			AllocatedStorage:     aws.Int32(allocatedStorageOrDefault(opt.AllocatedStorage)),
			DBName:               aws.String(dbName),

			MultiAZ:            aws.Bool(opt.MultiAZ),
			PubliclyAccessible: aws.Bool(opt.PubliclyAccessible),
		}

		// Optional networking
		if len(opt.SecurityGroupIDs) > 0 {
			createInput.VpcSecurityGroupIds = opt.SecurityGroupIDs
		}

		_, err = client.CreateDBInstance(ctx, createInput)
		if err != nil {
			helpers.Error("create DB instance failed: %v", err)
			return err
		}

		// Wait until the instance is available
		waiter := rds.NewDBInstanceAvailableWaiter(client)
		waitErr := waiter.Wait(ctx, &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(name),
		}, 30*time.Minute)
		if waitErr != nil {
			helpers.Error("waiting for DB instance to become available failed: %v", waitErr)
			return waitErr
		}

		describeOut, err = client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(name),
		})
		if err != nil || len(describeOut.DBInstances) == 0 {
			if err != nil {
				helpers.Error("describe DB instance after creation failed: %v", err)
			} else {
				helpers.Error("describe DB instance after creation returned no instances")
			}
			return fmt.Errorf("could not find DB instance after creation")
		}
		instance = &describeOut.DBInstances[0]
	}

	// 2) Get endpoint & port
	if instance.Endpoint == nil || instance.Endpoint.Address == nil {
		helpers.Error("DB instance %s does not have an endpoint yet", name)
		return fmt.Errorf("db instance has no endpoint")
	}

	host := aws.ToString(instance.Endpoint.Address)
	port := int(aws.ToInt32(instance.Endpoint.Port))
	if port == 0 {
		port = defaultPortForEngine(engine)
	}

	// DSN
	dsn := fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s",
		engine,
		username,
		password,
		host,
		port,
		dbName,
	)

	// 3) Export env vars
	helpers.Setenv("DB_ENGINE", engine)
	helpers.Setenv("DB_HOST", host)
	helpers.Setenv("DB_PORT", fmt.Sprintf("%d", port))
	helpers.Setenv("DB_NAME", dbName)
	helpers.Setenv("DB_USER", username)
	helpers.Setenv("DB_PASSWORD", password)
	helpers.Setenv("DB_DSN", dsn)

	helpers.Setenv("RDS_REGION", region)
	helpers.Setenv("RDS_ENDPOINT", fmt.Sprintf("%s:%d", host, port))
	helpers.Setenv("RDS_INSTANCE_IDENTIFIER", name)

	helpers.Info(
		"aws-compose-service (service=rds) ready for %s (engine=%s endpoint=%s:%d)",
		name,
		engine,
		host,
		port,
	)

	return nil
}

// RDSDown deletes the RDS instance (without final snapshot) for the given service.
func RDSDown(ctx context.Context, opt structs.Options) error {
	region := helpers.WithFallbackValue(opt.Region, "ap-southeast-1")
	name := helpers.WithFallbackValue(opt.Name, "rds")

	cfg, err := helpers.LoadAWSConfig(ctx, region)
	if err != nil {
		helpers.Error("unable to load AWS config: %v", err)
		return err
	}

	client := rds.NewFromConfig(cfg)

	helpers.Info("deleting RDS instance %s in region %s (skip final snapshot)", name, region)

	_, err = client.DeleteDBInstance(ctx, &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(name),
		SkipFinalSnapshot:    aws.Bool(true),
	})
	if err != nil {
		switch err.(type) {
		case *rdstypes.DBInstanceNotFoundFault:
			helpers.Info("RDS instance %s does not exist, nothing to delete", name)
			return nil
		default:
			// real error, proceed
		}

		helpers.Error("delete DB instance failed: %v", err)
		return err
	}

	// Wait until the instance is deleted
	waiter := rds.NewDBInstanceDeletedWaiter(client)
	waitErr := waiter.Wait(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(name),
	}, 30*time.Minute)
	if waitErr != nil {
		helpers.Error("waiting for DB instance to be deleted failed: %v", waitErr)
		return waitErr
	}

	helpers.Info("RDS instance %s successfully deleted", name)
	return nil
}
