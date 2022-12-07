package main

import (
	"github.com/disintegration/gift"
	"github.com/golang/groupcache"
	"github.com/pierrre/imageserver"
	imageserver_cache "github.com/pierrre/imageserver/cache"
	imageserver_cache_groupcache "github.com/pierrre/imageserver/cache/groupcache"
	imageserver_http "github.com/pierrre/imageserver/http"
	imageserver_http_crop "github.com/pierrre/imageserver/http/crop"
	imageserver_http_gift "github.com/pierrre/imageserver/http/gift"
	imageserver_http_image "github.com/pierrre/imageserver/http/image"
	imageserver_image "github.com/pierrre/imageserver/image"
	imageserver_image_crop "github.com/pierrre/imageserver/image/crop"
	imageserver_image_gamma "github.com/pierrre/imageserver/image/gamma"
	_ "github.com/pierrre/imageserver/image/gif"
	imageserver_image_gift "github.com/pierrre/imageserver/image/gift"
	_ "github.com/pierrre/imageserver/image/jpeg"
	_ "github.com/pierrre/imageserver/image/png"

	"github.com/cognusion/emery/hmac"

	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	groupcacheName = "ouchCache"
)

func startHTTPServer() {

	hmac.DebugOut = DebugOut

	http.Handle("/", http.StripPrefix("/", newImageHTTPHandler()))
	if config.GetString(ConfigHMACKey) != "" {
		hs := hmac.NewSigner(config.GetString(ConfigHMACKey), config.GetString(ConfigHMACSalt), config.GetDuration(ConfigHMACExpiration))

		http.Handle("/_sign/", http.StripPrefix("/_sign/", hs))
	}
	http.Handle("/favicon.ico", http.NotFoundHandler())
	initGroupcacheHTTPPool() // it automatically registers itself to "/_groupcache"
	http.HandleFunc("/stats", groupcacheStatsHTTPHandler)
	err := http.ListenAndServe(config.GetString(ConfigListen), nil)
	if err != nil {
		panic(err)
	}
}

func newImageHTTPHandler() http.Handler {
	return &imageserver_http.Handler{
		Parser: imageserver_http.ListParser([]imageserver_http.Parser{
			&imageserver_http.SourcePathParser{},
			&imageserver_http_crop.Parser{},
			&imageserver_http_gift.RotateParser{},
			&imageserver_http_gift.ResizeParser{},
			&imageserver_http_image.FormatParser{},
			&imageserver_http_image.QualityParser{},
			&hmac.Parser{},
		}),
		Server: newServer(),
	}
}

func newServer() imageserver.Server {
	srv, err := NewS3Server(config.GetString(ConfigAwsRegion), config.GetString(ConfigAwsAccessKey), config.GetString(ConfigAwsSecretKey), config.GetString(ConfigS3Bucket))
	if err != nil {
		panic(err)
	}
	srv = newServerImage(srv)
	srv = newServerGroupcache(srv)
	if config.GetString(ConfigHMACKey) != "" {
		srv = hmac.NewVerifier(srv, config.GetString(ConfigHMACKey), config.GetString(ConfigHMACSalt), config.GetDuration(ConfigHMACExpiration))
	}
	return srv
}

func newServerImage(srv imageserver.Server) imageserver.Server {
	return &imageserver.HandlerServer{
		Server: srv,
		Handler: &imageserver_image.Handler{
			Processor: imageserver_image_gamma.NewCorrectionProcessor(
				imageserver_image.ListProcessor([]imageserver_image.Processor{
					&imageserver_image_crop.Processor{},
					&imageserver_image_gift.RotateProcessor{
						DefaultInterpolation: gift.CubicInterpolation,
					},
					&imageserver_image_gift.ResizeProcessor{
						DefaultResampling: gift.LanczosResampling,
					},
				}), true)},
	}
}

func newServerGroupcache(srv imageserver.Server) imageserver.Server {
	if config.GetInt64(ConfigGroupCacheSize) <= 0 {
		// No groupcache for you
		return srv
	}
	return imageserver_cache_groupcache.NewServer(
		srv,
		imageserver_cache.NewParamsHashKeyGenerator(sha256.New),
		groupcacheName,
		config.GetInt64(ConfigGroupCacheSize),
	)
}

func initGroupcacheHTTPPool() {
	self := (&url.URL{Scheme: "http", Host: config.GetString(ConfigListen)}).String()
	var peers []string
	peers = append(peers, self)
	for _, p := range strings.Split(config.GetString(ConfigGroupCachePeers), ",") {
		if p == "" {
			continue
		}
		peer := (&url.URL{Scheme: "http", Host: p}).String()
		peers = append(peers, peer)
	}
	pool := groupcache.NewHTTPPool(self)
	pool.Context = imageserver_cache_groupcache.HTTPPoolContext
	pool.Transport = imageserver_cache_groupcache.NewHTTPPoolTransport(nil)
	pool.Set(peers...)
}

func groupcacheStatsHTTPHandler(w http.ResponseWriter, req *http.Request) {
	gp := groupcache.GetGroup(groupcacheName)
	if gp == nil {
		http.Error(w, fmt.Sprintf("group %s not found", groupcacheName), http.StatusServiceUnavailable)
		return
	}
	type cachesStats struct {
		Main groupcache.CacheStats
		Hot  groupcache.CacheStats
	}
	type stats struct {
		Group  groupcache.Stats
		Caches cachesStats
	}
	data, err := json.MarshalIndent(
		stats{
			Group: gp.Stats,
			Caches: cachesStats{
				Main: gp.CacheStats(groupcache.MainCache),
				Hot:  gp.CacheStats(groupcache.HotCache),
			},
		},
		"",
		"	",
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}
