package cri

import (
	"context"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"strings"
	"time"
)

type crioRootConfig struct {
	Crio struct {
		// Root is a path to the "root directory" where data not
		// explicitly handled by other options will be stored.
		Root string `toml:"root"`

		// RunRoot is a path to the "run directory" where state information not
		// explicitly handled by other options will be stored.
		RunRoot string `toml:"runroot"`
	} `toml:"crio"`
}

func fetchCriOVolumes(socketPath string) ([]string, error) {
	cli := &http.Client{Transport: &http.Transport{
		DisableCompression: true,
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, 32*time.Second)
		},
	}}

	req, err := http.NewRequest("GET", "/config", nil)
	if err != nil {
		klog.Errorf("unable to create http request: %s", err)
		return nil, err
	}

	req.Host = "crio"
	req.URL.Host = socketPath
	req.URL.Scheme = "http"

	resp, err := cli.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		klog.Errorf("unable to fetch cri-o configuration: %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("unable to read cri-o configuration response: %s", err)
		return nil, err
	}

	klog.Infof("cri-o configuration: %s", string(body))

	c := &crioRootConfig{}
	if _, err = toml.Decode(string(body), c); err != nil {
		klog.Errorf("unable to decode cri-o configuration: %s", err)
		return nil, err
	}

	if len(c.Crio.Root) == 0 || len(c.Crio.RunRoot) == 0 {
		klog.Error("invalid cri-o configuration")
		return nil, err
	}

	if strings.HasPrefix(c.Crio.Root, c.Crio.RunRoot) {
		return []string{c.Crio.RunRoot}, nil
	}

	if strings.HasPrefix(c.Crio.RunRoot, c.Crio.Root) {
		return []string{c.Crio.Root}, nil
	}

	return []string{c.Crio.Root, c.Crio.RunRoot}, nil
}
