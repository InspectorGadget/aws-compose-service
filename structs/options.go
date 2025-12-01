package structs

// Options holds all configuration passed from Docker Compose into the provider.
type Options struct {
	// Generic compose / provider metadata
	Project string
	Name    string

	// Service selector: rds, s3 (future: dynamodb, sqs, etc.)
	Service string

	// RDS-specific configuration
	Region        string
	Engine        string
	EngineVersion string

	InstanceClass    string
	AllocatedStorage int

	DBName   string
	Username string
	Password string

	// RDS networking
	SubnetIDs        []string
	SecurityGroupIDs []string

	PubliclyAccessible bool
	MultiAZ            bool

	// S3-specific configuration
	BucketName string
}
