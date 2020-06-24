package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"github.com/rcrowley/go-metrics"
	rpcx_client "github.com/smallnest/rpcx/client"
	rpcx_server "github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	nettools "github.com/toolkits/net"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net"
	"os"
	"os/signal"
	_runtime "runtime"
	"syscall"
	"time"
)

var (
	server    *rpcx_server.Server
	Config    Cfg
	localIP   string
	localName string
	//debug        = kingpin.Flag("debug", "Enable debug mode.").Bool()
	confFilePath = kingpin.Flag("config", "Provide a valid configuration path").Short('c').Default("./conf/").ExistingFileOrDir()
)

func init() {
	_runtime.GOMAXPROCS(_runtime.NumCPU())

	ips, err := nettools.IntranetIP()
	if err != nil {
		log.Fatalln(err)
	}
	if len(ips) == 0 {
		log.Fatalln("cant't get local ip")
	}
	localIP = ips[0]

	localName, err = os.Hostname()
	if err != nil {
		localName = localIP
	}
	kingpin.Version("v2.0")
	kingpin.Parse()

	err = ReadConfig("config", &Config, func(i interface{}) error {
		c := i.(*Cfg)
		if c.Mode == "" {
			c.Mode = "dev"
		}

		if c.Server.Name == "" {
			return errors.New("server's name is empty")
		}

		if err = Logger.init(); err != nil {
			return err
		}

		if c.Server.Proto == "" {
			c.Server.Proto = "tcp"
		}

		if c.Server.Host == "" {
			c.Server.Host = localIP
		}

		if c.Server.Port == 0 {
			l, _ := net.Listen("tcp", ":0")
			c.Server.Port = l.Addr().(*net.TCPAddr).Port
			l.Close()
		}

		if c.Server.ReadTimeout == 0 {
			c.Server.ReadTimeout = 3
		}
		if c.Server.WriteTimeout == 0 {
			c.Server.WriteTimeout = 3
		}

		if c.Register.Addr == "" {
			c.Register.Addr = "http://127.0.0.1:8500"
		}
		if c.Db != nil {
			if err := DB.Init(); err != nil {
				return err
			}
			Logger.Info("Db init success")
		}
		if c.Redis != nil {
			if err := Redis.Init(); err != nil {
				return err
			}
			Logger.Info("Redis init success")
		}
		if c.Mongo != nil {
			if err := Mongo.Init(); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		Logger.Fatal(err)
		return
	}

	server = rpcx_server.NewServer(
		rpcx_server.WithReadTimeout(time.Duration(Config.Server.ReadTimeout)*time.Second),
		rpcx_server.WithWriteTimeout(time.Duration(Config.Server.WriteTimeout)*time.Second),
	)

	if Config.Mode == "dev" {
		server.Plugins.Add(rpcx_client.InprocessClient)
	} else {
		r := &serverplugin.ConsulRegisterPlugin{
			ServiceAddress: "tcp@" + Config.Server.String(),
			ConsulServers:  []string{Config.Register.Addr},
			BasePath:       Config.Server.Name,
			Metrics:        metrics.NewRegistry(),
			UpdateInterval: time.Minute,
		}
		err = r.Start()
		if err != nil {
			Logger.Fatal(err)
		}
		server.Plugins.Add(r)
	}

	Logger.Info("bootstrap init success")
}

type Initialization func(*rpcx_server.Server) error

func Run(ctx context.Context, initFn ...Initialization) {
	for _, fn := range initFn {
		if err := fn(server); err != nil {
			Logger.Fatal(err)
			return
		}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	Logger.Info(fmt.Sprintf("server listen: %s", Config.Server.String()))
	go server.Serve(Config.Server.Proto, Config.Server.String())

	<-ch

	server.UnregisterAll()
	server.Shutdown(ctx)
	Logger.Info("server stopped")
}

func Register(name string, hdl interface{}) error {
	Logger.Info(fmt.Sprintf("register service: %s/%s", Config.Server.Name, name))
	return server.RegisterName(name, hdl, "")
}
