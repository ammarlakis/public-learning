/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

type listCmdInput struct {
	jsonOutput bool
	timeout    int
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List S3 buckets",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		timeout, _ := cmd.Flags().GetInt("timeout")
		listBuckets(&listCmdInput{
			jsonOutput,
			timeout,
		})
	},
}

func listBuckets(input *listCmdInput) {
	ctx, cancel := context.WithTimeoutCause(context.Background(), time.Duration(input.timeout)*time.Second, errors.New("Timeout"))

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Unable to load sdk config: %v", err)

	}

	s3client := s3.NewFromConfig(cfg)

	result, err := s3client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Fatalf("Request failed: %v", err)
		}
		log.Fatalf("Unable to list buckets: %v", err)
	}

	defer cancel()

	if input.jsonOutput {
		jsonData, _ := json.MarshalIndent(&result.Buckets, "", "  ")
		fmt.Println(string(jsonData))
	} else {
		fmt.Println("Buckets:")
		for _, bucket := range result.Buckets {
			fmt.Println(*bucket.Name)
		}
	}
}

func init() {
	listCmd.Flags().Bool("json", false, "Output in JSON format")
	listCmd.Flags().IntP("timeout", "t", 10, "Timeout in seconds")
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
