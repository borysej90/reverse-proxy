package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"proxy/internal/config"
	"proxy/internal/limiter"
)

func main() {
	configFile := "config.yaml"
	if envVar, ok := os.LookupEnv("CONFIG_FILE"); ok {
		configFile = envVar
	}
	file, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	cfg, err := config.FromYaml(file)
	file.Close()
	if err != nil {
		panic(err)
	}
	for _, path := range cfg.Server.Paths {
		log.Printf("Setting up forwarding for location %q to target %s", path.Location, path.Target)
		target, err := url.Parse(path.Target)
		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		http.Handle(path.Location, limiter.New(path.ConnectionLimit, proxy, path.DropOverLimit))
	}
	log.Printf("Starting a proxy at address :%s\n", cfg.Server.Listen)
	http.ListenAndServe(":"+cfg.Server.Listen, nil)
}
