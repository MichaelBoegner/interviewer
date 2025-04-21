# AWS Infrastructure Scripts

This folder contains AWS-specific deployment scripts for the Interviewer App.

## Files

- **deploy_to_aws.sh**  
  Automates ECR image push, ECS service update, and ALB listener validation.  
  Run this when deploying to AWS Fargate with an existing setup (cluster, task, ALB, etc.).

## Prerequisites

- ECS cluster and service already created
- ALB and Target Group already created
- SSL certificate issued in ACM
- Docker image already built and tagged

## Task Definition

This project uses a .gitignore preconfigured ECS task definition located at `infra/ecs/task-definition.json`. 

To register it manually:
```bash
aws ecs register-task-definition \
  --cli-input-json file://infra/ecs/task-definition.json
```

## Usage

```bash
chmod +x deploy_to_aws.sh
./deploy_to_aws.sh
```