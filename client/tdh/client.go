package tdh

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/model"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/auth"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/controller"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/core"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/customer-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/infra-connector"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/service-metadata"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/task"
	"github.com/svc-bot-mds/terraform-provider-tdh/client/tdh/upgrade-service"
	"net/http"
	"strings"
	"time"
)

// HostURL - Default TDH URL
const HostURL string = "http://localhost:8080"

// Client -
type Client struct {
	Root             *core.Root
	Auth             *auth.Service
	Controller       *controller.Service
	InfraConnector   *infra_connector.Service
	CustomerMetadata *customer_metadata.Service
	ServiceMetadata  *service_metadata.Service
	UpgradeService   *upgrade_service.Service
	TaskService      *task.Service
}

type TokenGetter func() (*auth.TokenResponse, error)

// NewClient -
func NewClient(host *string, authInfo *model.ClientAuth) (*Client, error) {
	hostUrl := HostURL
	if len(strings.TrimSpace(*host)) != 0 {
		hostUrl = *host
	}

	httpClient := prepareHttpClient()
	root := &core.Root{
		// Default TDH URL
		HostUrl:    &hostUrl,
		AuthToUse:  authInfo,
		HttpClient: httpClient,
	}

	c := prepareClient(host, root)
	root.TokenGetter = func() (any, error) {
		return c.Auth.GetAccessToken()
	}

	if err := c.Auth.Login(); err != nil {
		apiErr := core.ApiError{}
		if errors.As(err, &apiErr) {
			return nil, fmt.Errorf("%s", apiErr.ErrorMessage)
		}
		return nil, err
	}
	if _, err := c.Auth.GetAccessToken(); err != nil {
		return nil, err
	}

	return c, nil
}

func prepareClient(host *string, root *core.Root) *Client {
	return &Client{
		Root:             root,
		Auth:             auth.NewService(host, root),
		Controller:       controller.NewService(host, root),
		InfraConnector:   infra_connector.NewService(host, root),
		CustomerMetadata: customer_metadata.NewService(host, root),
		ServiceMetadata:  service_metadata.NewService(host, root),
		UpgradeService:   upgrade_service.NewService(host, root),
		TaskService:      task.NewService(host, root),
	}
}

func prepareHttpClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
