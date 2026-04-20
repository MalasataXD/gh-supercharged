package ghclient

import (
	"github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	REST *api.RESTClient
	GQL  *api.GraphQLClient
}

func New() (*Client, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}
	return &Client{REST: rest, GQL: gql}, nil
}
