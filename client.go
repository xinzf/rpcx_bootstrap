package bootstrap

import (
	"context"
	"errors"
	rpcx_client "github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
	"strings"
	"time"
)

type clientConfig struct {
	Name        string         `mapstructure:"name"`
	ReadTimeOut int64          `mapstructure:"read_timeout"`
	Compress    bool           `mapstructure:"compress"`
	Register    registerConfig `mapstructure:"register"`
}

type client struct {
	err   error
	meta  map[string]string
	async bool
}

var Client *client

func (this *client) Init() *client {
	if Config.Client.Name == "" {
		Config.Client.Name = "xunray.rpcx.client"
	}
	if Config.Register.Addr == "" {
		Config.Register.Addr = "http://127.0.0.1:8500"
	}
	if Config.Client.ReadTimeOut == 0 {
		Config.Client.ReadTimeOut = 5
	}

	Client = new(client)
	Client.meta = make(map[string]string)
	return Client
}

func (this *client) Async() *client {
	this.async = true
	return this
}

func (this *client) Meta(mp map[string]string) *client {
	this.meta = mp
	return this
}

func (this *client) callInProcess(svcPath, svcMethod string, req, rsp interface{}) error {
	d := rpcx_client.NewInprocessDiscovery()
	xclient := rpcx_client.NewXClient(svcPath, rpcx_client.Failtry, rpcx_client.RandomSelect, d, rpcx_client.DefaultOption)
	defer xclient.Close()

	err := xclient.Call(context.Background(), svcMethod, req, rsp)
	return err
}

func (this *client) callDiscovery(basePath, svcPath, svcMethod string, req, rsp interface{}) error {

	options := rpcx_client.DefaultOption
	options.ReadTimeout = time.Duration(Config.Client.ReadTimeOut) * time.Second
	if Config.Client.Compress {
		options.CompressType = protocol.Gzip
	}

	d := rpcx_client.NewConsulDiscoveryTemplate(basePath, []string{Config.Client.Register.Addr}, nil)
	oneClient := rpcx_client.NewOneClient(rpcx_client.Failover, rpcx_client.RandomSelect, d, options)
	defer oneClient.Close()

	metas := make(map[string]string)
	{
		metas["client"] = Config.Client.Name
		metas["remote_ip"] = localIP
		metas["remote_host"] = localName
		for k, v := range this.meta {
			metas[k] = v
		}
	}

	ctx := context.WithValue(context.Background(), share.ReqMetaDataKey, metas)

	if this.async {
		call, err := oneClient.Go(ctx, svcPath, svcMethod, req, rsp, nil)
		if err != nil {
			return err
		}

		replyCall := <-call.Done
		if replyCall.Error != nil {
			return replyCall.Error
		}
		return nil
	}
	return oneClient.Call(ctx, svcPath, svcMethod, req, rsp)
}

func (this *client) Call(name string, req, rsp interface{}) error {
	if this.err != nil {
		return this.err
	}

	strs := strings.Split(name, ".")
	if len(strs) < 3 {
		return errors.New("call service name is unvalid")
	}

	basePath := strs[0]
	svcName := strs[1]
	methodName := strings.Join(strs[2:], ".")

	if basePath == "" || svcName == "" || methodName == "" {
		return errors.New("call service name is unvalid")
	}

	var err error
	if Config.Mode == "dev" {
		err = this.callInProcess(svcName, methodName, req, rsp)
	} else {
		err = this.callDiscovery(basePath, svcName, methodName, req, rsp)
	}

	return err
}
