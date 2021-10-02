package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Using the Config value, create the DynamoDB client
	svc := ec2.NewFromConfig(cfg)

	// Build the request with its input parameters
	resp, err := svc.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	// resp, err := svc.ListTables(context.TODO(), &dynamodb.ListTablesInput{
	// 	Limit: aws.Int32(5),
	// })
	if err != nil {
		log.Fatalf("failed to list tables, %v", err)
	}

	// fmt.Printf("%+v\n", resp)
	// spew.Dump(resp.Vpcs)
	// fmt.Println("VPCs: ", resp)
	for _, vpc := range resp.Vpcs {
		fmt.Println(*vpc.VpcId)
	}
}
