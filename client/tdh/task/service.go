package task

import (
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"strings"
)

var (
	defaultPage = &model.PageQuery{
		Index: 0,
		Size:  100,
		Sort:  "modified,DESC",
	}
)

const (
	EndPoint = "task"
)

type Service struct {
	*core.Service
}

func NewService(hostUrl *string, root *core.Root) *Service {
	return &Service{
		Service: core.NewService(hostUrl, EndPoint, root),
	}
}

// GetTasks - Returns page of tasks
func (s *Service) GetTasks(query *TasksQuery) (model.Paged[model.Task], error) {
	urlPath := fmt.Sprintf("%s/%s", s.Endpoint, Tasks)
	var response model.Paged[model.Task]

	if query.Size == 0 {
		query.Size = defaultPage.Size
	}

	_, err := s.Api.Get(&urlPath, query, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// GetTask - Return dto of task
func (s *Service) GetTask(id string) (*model.Task, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}
	urlPath := fmt.Sprintf("%s/%s/%s/%s", s.Endpoint, Tasks, Info, id)
	var response model.Task

	_, err := s.Api.Get(&urlPath, nil, &response)
	if err != nil {
		return &response, err
	}

	return &response, nil
}
