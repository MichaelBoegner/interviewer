#!/bin/bash

set -e

# CONFIGURATION
AWS_REGION="us-east-1"
CLUSTER_NAME="interviewer-cluster"
SERVICE_NAME="interviewer-service"
TARGET_GROUP_ARN="arn:aws:elasticloadbalancing:us-east-1:742087940360:targetgroup/interviewer-target-group/453a2f90e4f2c535"
ALB_ARN="arn:aws:elasticloadbalancing:us-east-1:742087940360:loadbalancer/app/interviewer-alb/a131c9475c4cb773"
CERT_ARN="arn:aws:acm:us-east-1:742087940360:certificate/a1d8882d-83a9-4dd8-a631-420cf8ecf170"
ECR_IMAGE="742087940360.dkr.ecr.us-east-1.amazonaws.com/interviewer:latest"

echo "🔁 Building Docker image..."
docker build --no-cache -t interviewer .

echo "🔐 Logging into ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin 742087940360.dkr.ecr.$AWS_REGION.amazonaws.com

echo "📦 Tagging and pushing image to ECR..."
docker tag interviewer:latest $ECR_IMAGE
docker push $ECR_IMAGE

echo "🚀 Forcing new ECS deployment..."
aws ecs update-service \
  --cluster $CLUSTER_NAME \
  --service $SERVICE_NAME \
  --force-new-deployment \
  --region $AWS_REGION

echo "✅ ECS deployment triggered. Now ensuring HTTPS listener is present..."

# Check if HTTPS listener already exists
EXISTING_LISTENER=$(aws elbv2 describe-listeners \
  --load-balancer-arn $ALB_ARN \
  --region $AWS_REGION \
  --query "Listeners[?Port==\`443\`].ListenerArn" \
  --output text)

if [ -z "$EXISTING_LISTENER" ]; then
  echo "🔒 Creating HTTPS listener on port 443..."
  aws elbv2 create-listener \
    --load-balancer-arn $ALB_ARN \
    --protocol HTTPS \
    --port 443 \
    --certificates CertificateArn=$CERT_ARN \
    --default-actions Type=forward,TargetGroupArn=$TARGET_GROUP_ARN \
    --region $AWS_REGION
else
  echo "✅ HTTPS listener already exists at ARN:"
  echo $EXISTING_LISTENER
fi

echo "🎉 Deployment script complete."
