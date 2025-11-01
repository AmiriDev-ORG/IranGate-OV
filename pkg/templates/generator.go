package templates
import (
	"bytes"
	"text/template"
)
type ConfigGenerator struct {
	serverTemplate *template.Template
	clientTemplate *template.Template
}
func NewConfigGenerator() (*ConfigGenerator, error) {
	serverTmpl, err := template.New("server").Parse(DefaultServerConfig)
	if err != nil {
		return nil, err
	}
	clientTmpl, err := template.New("client").Parse(DefaultClientConfig)
	if err != nil {
		return nil, err
	}
	return &ConfigGenerator{
		serverTemplate: serverTmpl,
		clientTemplate: clientTmpl,
	}, nil
}
func (g *ConfigGenerator) GenerateServerConfig(settings map[string]interface{}) (string, error) {
	mergedSettings := make(map[string]interface{})
	for k, v := range DefaultServerSettings {
		mergedSettings[k] = v
	}
	for k, v := range settings {
		mergedSettings[k] = v
	}
	var buf bytes.Buffer
	if err := g.serverTemplate.Execute(&buf, mergedSettings); err != nil {
		return "", err
	}
	return buf.String(), nil
}
func (g *ConfigGenerator) GenerateClientConfig(settings map[string]interface{}) (string, error) {
	mergedSettings := make(map[string]interface{})
	for k, v := range DefaultClientSettings {
		mergedSettings[k] = v
	}
	for k, v := range settings {
		mergedSettings[k] = v
	}
	var buf bytes.Buffer
	if err := g.clientTemplate.Execute(&buf, mergedSettings); err != nil {
		return "", err
	}
	return buf.String(), nil
}