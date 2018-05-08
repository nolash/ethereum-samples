package resource

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/log"

	"../protocol"
)

type Client struct {
	url      string
	resource string
	client   *http.Client
	ready    bool
}

func NewClient(bzzapi string, resource string) *Client {
	return &Client{
		client:   http.DefaultClient,
		resource: resource,
		url:      bzzapi,
	}
}
func (b *Client) createResource(data []byte) error {
	_, err := b.client.Post(
		fmt.Sprintf("%s/bzz-resource:/%s/raw/2", b.url, b.resource),
		"contenxt-type: application/octet-stream",
		bytes.NewBuffer(data),
	)
	if err == nil {
		log.Debug("creating resource", "id", b.resource)
		b.ready = true
	}
	return err
}

func (b *Client) updateResource(data []byte) error {
	if !b.ready {
		return b.createResource(data)
	}
	_, err := b.client.Post(
		fmt.Sprintf("%s/bzz-resource:/%s/raw", b.url, b.resource),
		"content-type: application/octet-stream",
		bytes.NewBuffer(data),
	)
	return err
}

func (b *Client) ResourceSinkFunc() func(interface{}) {
	return func(obj interface{}) {
		if res, ok := obj.(*protocol.Result); ok {
			log.Debug("posting", "obj", fmt.Sprintf("%x", res.Hash))
			if err := b.updateResource(res.Hash); err != nil {
				log.Error("resource fail", "err", err, "hash", res.Hash)
			}
		}
	}
}
func main() {
	fmt.Println("vim-go")
}
