# aws-compose-service

> Docker Compose AWS provider for **RDS** and **S3**

This project implements a **Docker Compose extension / provider** that wires
AWS services into your Compose applications by:

- Creating or reusing **RDS** databases
- Creating or reusing **S3** buckets
- Emitting **JSONL** messages and **environment variables** that Docker Compose
  understands via the provider protocol.

It is designed to be used from `docker-compose.yml` as an extension, but it can
also be run directly from the CLI for debugging.

---
### Feature: RDS

- Creates an RDS instance if it does not exist, or reuses an existing one:
  - `CreateDBInstance`
  - Waits until instance is **available**
- Exports connection details as environment variables:
  - `DB_ENGINE`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_DSN`
  - `RDS_REGION`, `RDS_ENDPOINT`, `RDS_INSTANCE_IDENTIFIER`

Available options for Compose:
| Option                | Type   | Required | Description                                   |
| --------------------- | ------ | -------- | --------------------------------------------- |
| `service`             | string | yes      | Must be `rds`                                 |
| `region`              | string | no       | AWS region (default: `ap-southeast-1`)        |
| `name`                | string | no       | Instance identifier (default: Compose `name`) |
| `engine`              | string | yes      | `postgres`, `mysql`, `mariadb`, `sqlserver`   |
| `engine_version`      | string | no       | Engine version                                |
| `db_name`             | string | no       | Default: `app`                                |
| `username`            | string | yes      | Master user                                   |
| `password`            | string | yes      | Master password                               |
| `instance_class`      | string | no       | Default: `db.t3.micro`                        |
| `allocated_storage`   | int    | no       | Default: `20` GiB                             |
| `publicly_accessible` | bool   | no       | Default: `false`                              |
| `multi_az`            | bool   | no       | Default: `false`                              |
| `subnet_ids`          | list   | no       | Optional subnet list                          |
| `security_group_ids`  | list   | no       | Optional SG list                              |
| `project`             | string | auto     | Provided by Compose                           |
| `name`                | string | auto     | Provided by Compose                           |

---
### Feature: S3

- Ensures an S3 bucket exists:
  - If the bucket is missing, calls `CreateBucket`
  - If it exists, reuses it
- Exports bucket details as environment variables:
  - `BUCKET_NAME`, `BUCKET_REGION`, `BUCKET_URL`
  - `S3_BUCKET_NAME`, `S3_BUCKET_REGION`, `S3_BUCKET_URL`

Available options for Compose:
| Option        | Type   | Required | Description                                                  |
| ------------- | ------ | -------- | ------------------------------------------------------------ |
| `service`     | string | yes      | Must be `s3`                                                 |
| `region`      | string | no       | AWS region (default: `ap-southeast-1`)                       |
| `bucket_name` | string | no       | If not provided, auto-generated: `<project>-<name>-<region>` |
| `name`        | string | no       | Logical name from Docker Compose                             |
| `project`     | string | no       | Project name (provided automatically by Docker Compose)      |

---
### JSONL Protocol

All communication back to Docker Compose is done using **JSON Lines**, e.g.:

```json
{"type":"info","message":"aws-compose-service (service=rds) ready for api (engine=postgres endpoint=...)"}
{"type":"setenv","message":"DB_HOST=my-rds-instance.abc123.ap-southeast-1.rds.amazonaws.com"}
```

---
### Requirements
- Go 1.22+
- Valid AWS credentials (one of):
    - `aws configure` default profile
    - `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables
    - IAM Role (e.g. in EC2, ECS, etc)
- IAM permissions for RDS and S3 operations
    - For RDS: `rds:DescribeDBInstances`, `rds:CreateDBInstance`, `rds:DeleteDBInstance`, etc.
    - For S3: `s3:CreateBucket`, `s3:DeleteBucket`, `s3:ListBucket`, etc.

---
### Building
```
git clone https://github.com/InspectorGadget/aws-compose-service.git
cd aws-compose-service

go mod tidy
go build -o aws-compose-service .
```

---
### CLI Usage (Debugging)
RDS Up
```
./aws-compose-service \
  --project myapp \
  --name api-db \
  --service rds \
  up \
  --region ap-southeast-1 \
  --engine postgres \
  --db_name myapp \
  --username appuser \
  --password supersecret
```

Example JSONL output:
```json
{"type":"info","message":"creating RDS instance api-db in ap-southeast-1 (engine=postgres)"}
{"type":"setenv","message":"DB_HOST=api-db.abc123.ap-southeast-1.rds.amazonaws.com"}
{"type":"setenv","message":"DB_PORT=5432"}
{"type":"setenv","message":"DB_NAME=myapp"}
{"type":"setenv","message":"DB_USER=appuser"}
{"type":"setenv","message":"DB_PASSWORD=supersecret"}
{"type":"setenv","message":"DB_DSN=postgres://appuser:supersecret@api-db.abc123.ap-southeast-1.rds.amazonaws.com:5432/myapp"}
{"type":"info","message":"aws-compose-service (service=rds) ready for api-db (engine=postgres endpoint=...)"}
```

S3 Up
```
./aws-compose-service \
  --project myapp \
  --name assets \
  --service s3 \
  up \
  --region ap-southeast-1 \
  --bucket_name myapp-assets-prod
```

Example JSONL output:
```json
{"type":"info","message":"ensuring S3 bucket myapp-assets-prod exists in ap-southeast-1"}
{"type":"setenv","message":"BUCKET_NAME=myapp-assets-prod"}
{"type":"setenv","message":"BUCKET_REGION=ap-southeast-1"}
{"type":"setenv","message":"BUCKET_URL=https://myapp-assets-prod.s3.ap-southeast-1.amazonaws.com/"}
{"type":"info","message":"aws-compose-service (service=s3) ready for assets (bucket=myapp-assets-prod)"}
```

---
### Example: Docker Compose Extension

```yaml
services:
  # database:
  #   provider:
  #     type: aws-compose-service
  #     options:
  #       service: rds
  #       region: ap-southeast-1
  #       name: mydbinstance
  #       engine: postgres
  #       engine_version: "13.4"
  #       instance_class: db.t3.micro
  #       allocated_storage: 20
  #       username: admin
  #       password: password123
  #       subnet_ids:
  #         - subnet-0bb1c79de3EXAMPLE
  #         - subnet-064f5c1e2fEXAMPLE
  #       security_group_ids:
  #         - sg-0a1b2c3d4e5f6g78h
  s3:
    provider:
      type: aws-compose-service
      options:
        service: s3
        region: ap-southeast-1

  app:
    image: alpine:3.22.2
    depends_on:
      - s3
    command: ["sh", "-c", "env | sort && sleep 3600"]
```