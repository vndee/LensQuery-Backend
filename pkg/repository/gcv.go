package repository

import (
	"context"

	vision "cloud.google.com/go/vision/apiv1"
)

type IGCVClient struct {
	client *vision.ImageAnnotatorClient
}

var GCVClient = &IGCVClient{}

func (g *IGCVClient) Init() error {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return err
	}
	g.client = client
	return nil
}
