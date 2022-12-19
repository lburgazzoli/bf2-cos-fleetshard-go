package camel

type ConnectorType string

const (
	ConnectorTypeSource ConnectorType = "source"
	ConnectorTypeSink   ConnectorType = "sink"
)

type ServiceAccount struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type Operator struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type EndpointKamelet struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

type Kamelets struct {
	Adapter     EndpointKamelet   `json:"adapter"`
	Kafka       EndpointKamelet   `json:"kafka"`
	Annotations map[string]string `json:"processors,omitempty"`
}

type ShardMetadata struct {
	ConnectorImage       string            `json:"connector_image"`
	ConnectorType        ConnectorType     `json:"connector_type"`
	Annotations          map[string]string `json:"annotations,omitempty"`
	Operators            []Operator        `json:"operators,omitempty"`
	Kamelets             Kamelets          `json:"kamelets"`
	Consumes             string            `json:"consumes"`
	ConsumesClass        string            `json:"consumes_class"`
	Produces             string            `json:"produces"`
	ProducesClass        string            `json:"produces_class"`
	ErrorHandlerStrategy string            `json:"error_handler_strategy"`
}
