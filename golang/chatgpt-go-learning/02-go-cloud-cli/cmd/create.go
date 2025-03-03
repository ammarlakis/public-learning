/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creating a new bucket",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		bucketName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Failed to get bucket name: %v", err)
		}
		createBucket(bucketName)
	},
}

func createBucket(name string) {
	timeoutErr := errors.New("Timeout")

	ctx, cancel := context.WithTimeoutCause(context.Background(), 10*time.Second, timeoutErr)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	s3client := s3.NewFromConfig(cfg)

	_, err = s3client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &name,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("Request failed: %v", err)
		}
		log.Fatalf("Failed to create bucket: %v", err)
	}

	fmt.Printf("Bucket %s created successfully", name)
}

func init() {
	createCmd.Flags().StringP("name", "n", "", "Name of the S3 bucket")
	createCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
