package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	qovery "github.com/qovery/qovery-client-go"
	log "github.com/sirupsen/logrus"
	"http-handler-to-redeploy/src"
	"strings"
	"time"
)

type QueueData struct {
	Webhook src.PrismicWebhook
	Context gin.Context
}

func redeployHandler(q *src.Queue[QueueData]) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		webhook := src.PrismicWebhook{}
		if err := c.BindJSON(&webhook); err != nil {
			c.JSON(400, gin.H{
				"status": "error",
				"error":  err.Error(),
				"body":   c.Request.Body,
			})
			return
		}

		_ = q.Enqueue(QueueData{
			Webhook: webhook,
			Context: *c.Copy(), // copy gin context since it's not thread safe
		})

		c.JSON(200, gin.H{
			"status": "ok",
		})
	}

	return fn
}

func redeploy(c *gin.Context, client *qovery.APIClient, environmentId string, applicationId string) {
	for {
		// check environment status before trying to redeploy; otherwise it will fail
		status, _, err := client.EnvironmentMainCallsApi.GetEnvironmentStatus(context.Background(), environmentId).Execute()
		if err != nil {
			log.Errorf("error while getting environment status: %v", err)
			return
		}

		if isTerminalState(status.State) {
			log.Infof("environment %s is in a terminal state, continuing", environmentId)
			break
		}

		log.Infof("environment %s is not in a terminal state, waiting 5 seconds before trying again", environmentId)
		time.Sleep(5 * time.Second)
	}

	// change environment variable to force app rebuild
	envVars, _, err := client.ApplicationEnvironmentVariableApi.ListApplicationEnvironmentVariable(context.Background(), applicationId).Execute()
	if err != nil {
		log.Errorf("error while getting environment variables: %v", err)
		return
	}

	envVarKey := "TO_REDEPLOY_FAKE"
	envVar := findEnvironmentVariable(envVars.Results, envVarKey)

	if envVar == nil {
		log.Errorf("error while getting environment variable: %s", envVarKey)
		return
	}

	_, _, err = client.ApplicationEnvironmentVariableApi.EditApplicationEnvironmentVariable(context.Background(), applicationId, envVar.Id).
		EnvironmentVariableEditRequest(qovery.EnvironmentVariableEditRequest{
			Key:   envVar.Key,
			Value: fmt.Sprintf("%d", time.Now().Unix()),
		}).Execute()

	if err != nil {
		log.Errorf("error while editing environment variable: %v", err)
		return
	}

	// Redeploy with Qovery
	_, _, err = client.ApplicationActionsApi.RedeployApplication(context.Background(), applicationId).Execute()

	if err != nil {
		log.Errorf("error while redeploying application: %v", err)
		return
	}
}

func isTerminalState(state qovery.StateEnum) bool {
	return state == qovery.STATEENUM_RUNNING || state == qovery.STATEENUM_DELETED ||
		state == qovery.STATEENUM_STOPPED || state == qovery.STATEENUM_CANCELED ||
		state == qovery.STATEENUM_READY || strings.HasSuffix(string(state), "ERROR")
}

func findEnvironmentVariable(envVars []qovery.EnvironmentVariable, key string) *qovery.EnvironmentVariable {
	for _, envVar := range envVars {
		if envVar.Key == key {
			return &envVar
		}
	}

	return nil
}

func processQueue(q *src.Queue[QueueData], config *src.Config) {
	conf := qovery.NewConfiguration()
	conf.DefaultHeader["Authorization"] = "Token " + strings.TrimSpace(config.QoveryApiToken)
	client := qovery.NewAPIClient(conf)

	app, _, err := client.ApplicationMainCallsApi.GetApplication(context.Background(), config.QoveryProdApplicationId).Execute()

	if err != nil {
		panic(err)
	}

	environmentId := app.Environment.Id

	for {
		if queueData, ok := q.Dequeue(); ok {
			redeploy(&queueData.Context, client, environmentId, config.QoveryProdApplicationId)
			// TODO: add support for staging
			/*var tags []src.Tag
			tags = append(tags, queueData.Webhook.Tags.Addition...)
			tags = append(tags, queueData.Webhook.Tags.Deletion...)

			for _, tag := range tags {
				switch strings.ToLower(tag.ID) {
				case "production":
					// redeploy production
					redeploy(&queueData.Context, client, environmentId, config.QoveryProdApplicationId)
				case "staging":
					// redeploy staging
					redeploy(&queueData.Context, client, environmentId, config.QoveryStagingApplicationId)
				}
			}*/
		}
	}
}

func main() {
	config := src.LoadConfig()

	q := src.NewQueue[QueueData](1000, true)

	go func(q *src.Queue[QueueData], config *src.Config) {
		processQueue(q, config)
	}(q, &config)

	r := gin.Default()
	//r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.POST("/handler", redeployHandler(q))

	err := r.Run(fmt.Sprintf(":%d", config.HttpPort))

	if err != nil {
		return
	}
}
