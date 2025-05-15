package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/config"
	"github.com/georgemblack/bluesky-lonely-posts/pkg/util"
)

const (
	ECSClusterName = "bluesky"
	ECSServiceName = "server"
	DNSRecordName  = "feedgen.george.black"
)

// Query AWS APIs to find the public IP address of the most recent ECS server task.
// Update the DNS record in Cloudflare to use the new IP address.
// This is a hack to save money; the proper solution would involve using a load balancer.
func updateServiceDNS(config config.Config) error {
	// Create AWS clients
	cfg, err := awsConfig.LoadDefaultConfig(context.Background(), awsConfig.WithRegion("us-east-2"))
	if err != nil {
		return util.WrapErr("failed to load aws config", err)
	}
	ecsClient := ecs.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	//
	taskARNs, err := ecsClient.ListTasks(context.Background(), &ecs.ListTasksInput{
		Cluster:     aws.String(ECSClusterName),
		ServiceName: aws.String(ECSServiceName),
	})
	if err != nil {
		return util.WrapErr("failed to list ecs tasks", err)
	}
	if len(taskARNs.TaskArns) == 0 {
		return errors.New("no ecs tasks found")
	}

	tasks, err := ecsClient.DescribeTasks(context.Background(), &ecs.DescribeTasksInput{
		Cluster: aws.String(ECSClusterName),
		Tasks:   taskARNs.TaskArns,
	})
	if err != nil {
		return util.WrapErr("failed to describe ecs tasks", err)
	}
	if len(tasks.Tasks) == 0 {
		return errors.New("no ecs tasks found")
	}

	// Find the ECS task started most recently
	sort.Slice(tasks.Tasks, func(i, j int) bool {
		return tasks.Tasks[i].StartedAt.After(*tasks.Tasks[j].StartedAt)
	})

	newestTask := tasks.Tasks[0]
	eniID := ""
	for _, attachment := range newestTask.Attachments {
		for _, detail := range attachment.Details {
			if detail.Name != nil && *detail.Name == "networkInterfaceId" {
				eniID = *detail.Value
				break
			}
		}
	}

	if eniID == "" {
		return errors.New("failed to find eni id in task attachments")
	}

	// Fetch public IP address of the ENI
	eni, err := ec2Client.DescribeNetworkInterfaces(context.Background(), &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{eniID},
	})
	if err != nil {
		return util.WrapErr("failed to describe network interface", err)
	}
	if len(eni.NetworkInterfaces) == 0 || eni.NetworkInterfaces[0].Association == nil || eni.NetworkInterfaces[0].Association.PublicIp == nil {
		return errors.New("failed to find public ip address in network interface")
	}
	publicIP := *eni.NetworkInterfaces[0].Association.PublicIp
	slog.Info("found public ip address", "ip", publicIP)

	// Update DNS record in Cloudflare
	recordID := getDNSRecordID(DNSRecordName, config.CloudflareZoneID, config.CloudflareAPIToken)
	if recordID == "" {
		return errors.New("failed to find existing dns record")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"type":    "A",
		"name":    DNSRecordName,
		"content": publicIP,
		"ttl":     1,
		"proxied": true,
		"comment": "Managed by bluesky-lonely-posts",
	})
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.CloudflareZoneID, recordID)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+config.CloudflareAPIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("failed to update DNS record: %w", err))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to update dns record: " + resp.Status)
	}
	defer resp.Body.Close()

	return nil
}

func getDNSRecordID(name, zoneID, token string) string {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", zoneID, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Error(util.WrapErr("failed to create request", err).Error())
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(util.WrapErr("failed to send request", err).Error())
		return ""
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("failed to get dns record: " + resp.Status)
		return ""
	}
	defer resp.Body.Close()

	var data struct {
		Result []struct {
			ID string `json:"id"`
		} `json:"result"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(util.WrapErr("failed to read response body", err).Error())
		return ""
	}
	if err := json.Unmarshal(body, &data); err != nil {
		slog.Error(util.WrapErr("failed to unmarshal response body", err).Error())
		return ""
	}
	if len(data.Result) == 0 {
		slog.Error("no dns records found")
		return ""
	}

	return data.Result[0].ID
}
