package v_3_5_2

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"text/template"

	"strings"

	"github.com/giantswarm/microerror"
)

const (
	defaultRegistryDomain = "quay.io"
)

type CloudConfigConfig struct {
	Params   Params
	Template string
}

func DefaultCloudConfigConfig() CloudConfigConfig {
	return CloudConfigConfig{
		Params:   Params{},
		Template: "",
	}
}

type CloudConfig struct {
	config   string
	params   Params
	template string
}

func NewCloudConfig(config CloudConfigConfig) (*CloudConfig, error) {
	if err := config.Params.Validate(); err != nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Params.%s", err)
	}
	if config.Template == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Template must not be empty")
	}

	// Default to 443 for non AWS providers.
	if config.Params.EtcdPort == 0 {
		config.Params.EtcdPort = 443
	}
	// Set default registry to quay.io
	if config.Params.RegistryDomain == "" {
		config.Params.RegistryDomain = defaultRegistryDomain
	}
	// extract cluster base domain
	config.Params.BaseDomain = strings.TrimPrefix(config.Params.Cluster.Kubernetes.API.Domain, "api.")

	c := &CloudConfig{
		config:   "",
		params:   config.Params,
		template: config.Template,
	}

	return c, nil
}

func (c *CloudConfig) ExecuteTemplate() error {
	tmpl, err := template.New("cloudconfig").Parse(c.template)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, c.params)
	if err != nil {
		return err
	}
	c.config = buf.String()

	return nil
}

func (c *CloudConfig) Base64() string {
	cloudConfigBytes := []byte(c.config)

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(cloudConfigBytes)
	w.Close()

	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func (c *CloudConfig) String() string {
	return c.config
}
