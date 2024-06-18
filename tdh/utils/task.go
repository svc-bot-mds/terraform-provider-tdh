package utils

import (
	"context"
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"sync"
	"time"
)

func WaitForTask(client *tdh.Client, taskId string) error {
	for true {
		taskResponse, err := client.TaskService.GetTask(taskId)
		if err != nil {
			return err
		}
		if taskResponse.Status == "SUCCESS" {
			return nil
		} else if taskResponse.Status == "FAILED" {
			return fmt.Errorf("task has failed, get more details using datasource \"tdh_tasks\"")
		}
		time.Sleep(time.Second * 10)
	}
	return nil
}

func WaitForTaskV2(client *tdh.Client, taskId string) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Add(1)
	pollingChan := make(chan error)
	go func(channel chan error) {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Second)
		for {
			taskResponse, err := client.TaskService.GetTask(taskId)
			if err != nil {
				channel <- err
				return
			}
			if taskResponse.Status == "SUCCESS" {
				channel <- nil
				return
			} else if taskResponse.Status == "FAILED" {
				channel <- fmt.Errorf("task has failed, get more details using datasource \"tdh_tasks\"")
				return
			}
			select {
			case <-ctx.Done():
				fmt.Println("cancelled")
				return
			case <-ticker.C:
			}
		}
	}(pollingChan)
	wg.Wait()
	return <-pollingChan
}
