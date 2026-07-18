package aws

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetAPIKey() string {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey != "" {
		return apiKey
	}

	fmt.Println("🛡️  Sentinel Agent: Local AI_API_KEY not found. Fallback: Querying AWS Secrets Manager...")

	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("⚠️  AWS Config Error: %v", err)
		return ""
	}

	client := secretsmanager.NewFromConfig(cfg)
	secretName := "sentinel/gemini-api-key"

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(ctx, input)
	if err != nil {
		log.Printf("❌ AWS Secrets Manager Error: %v", err)
		return ""
	}

	if result.SecretString != nil {
		return *result.SecretString
	}

	return ""
}
