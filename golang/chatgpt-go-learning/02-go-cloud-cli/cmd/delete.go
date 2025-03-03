/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

func deleteBucket(name string) {
	timeoutErr := errors.New("Timeout")
	ctx, cancel := context.WithTimeoutCause(context.Background(), 5*time.Second, timeoutErr)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Error loading default config: %v", err)
	}

	s3client := s3.NewFromConfig(cfg)

	_, err = s3client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &name,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("Request failed: %v", err)
		}
		log.Fatalf("Error deleting bucket: %v", err)
	}

	log.Printf("Bucket %s deleted successfully", name)

}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		bucketName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Error reading flags: %v", err)
		}
		deleteBucket(bucketName)
	},
}

func init() {
	deleteCmd.Flags().StringP("name", "n", "", "Bucket Name")
	deleteCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
