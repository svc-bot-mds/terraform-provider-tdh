package utils

import (
	"context"
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh"
	"sync"
	"time"
)

type taskWaitResponse struct {
	TaskId string `json:"taskId"`
	Error  error  `json:"error"`
}

func WaitForTask(client *tdh.Client, taskId string) error {
	for true {
		taskResponse, err := client.TaskService.GetTask(taskId)
		if err != nil {
			return err
		}
		if taskResponse.Status == "SUCCESS" {
			return nil
		} else if taskResponse.Status == "FAILED" {
			return fmt.Errorf("task [ID: %s] has failed, get more details using datasource \"tdh_tasks\"", taskId)
		}
		time.Sleep(time.Second * 10)
	}
	return nil
}

func WaitForTaskV2(client *tdh.Client, taskId string, superChan *chan taskWaitResponse, wg *sync.WaitGroup) chan error {
	ch := make(chan error, 1)
	sendIt := func(taskId string, err error) {
		if superChan == nil {
			ch <- err
		} else {
			*superChan <- taskWaitResponse{taskId, err}
		}
	}
	go func(sendIt func(taskId string, err error)) {
		for true {
			taskResponse, err := client.TaskService.GetTask(taskId)
			if err != nil {
				sendIt(taskId, err)
				break
			}
			if taskResponse.Status == "SUCCESS" {
				sendIt(taskId, nil)
				break
			} else if taskResponse.Status == "FAILED" {
				sendIt(taskId, fmt.Errorf("task [ID: %s] has failed, get more details using datasource \"tdh_tasks\"", taskId))
				break
			}
			time.Sleep(time.Second * 10)
		}
		if wg != nil {
			wg.Done()
		}
	}(sendIt)
	return ch
}

func WaitForAllTasks(client *tdh.Client, taskResponseList []model.TaskResponse) error {
	if len(taskResponseList) == 0 {
		return nil
	}
	var taskIds []string
	for _, taskResponse := range taskResponseList {
		taskIds = append(taskIds, taskResponse.TaskId)
	}
	totalTasks := len(taskIds)
	bokaChan := make(chan taskWaitResponse, totalTasks)
	wg := sync.WaitGroup{}
	for _, taskId := range taskIds {
		wg.Add(1)
		go WaitForTaskV2(client, taskId, &bokaChan, &wg)
	}

	// now we wait for everyone to finish - again, not a must.
	// you can just receive from the channel N times, and use a timeout or something for safety
	wg.Wait()

	// we need to close the channel or the following loop will get stuck
	close(bokaChan)

	var failedTaskIds []string
	for response := range bokaChan {
		if response.Error != nil {
			taskIds = append(taskIds, response.TaskId)
		}
	}
	if len(failedTaskIds) > 0 {
		return fmt.Errorf("some tasks have failed %q, get more details using datasource \"tdh_tasks\"", failedTaskIds)
	}
	return nil
}

func WaitForTaskV3(client *tdh.Client, taskId string) error {
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
