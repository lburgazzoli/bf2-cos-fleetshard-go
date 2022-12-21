package camel

import (
	"fmt"

	camelv1alpha1 "github.com/apache/camel-k/pkg/apis/camel/v1alpha1"

	"github.com/pkg/errors"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/camel/endpoints"
)

func configureSteps(binding *camelv1alpha1.KameletBinding, cc ConnectorConfiguration) error {

	//
	// Decoder
	//

	decoder, err := configureDecoder(cc)
	if err != nil {
		return errors.Wrap(err, "error creating decoder")
	}
	if decoder != nil {
		binding.Spec.Steps = append(binding.Spec.Steps, *decoder)
	}

	//
	// Processors
	//
	// TODO

	//
	// Encoder
	//

	encoder, err := configureEncoder(cc)
	if err != nil {
		return errors.Wrap(err, "error creating encoder")
	}
	if encoder != nil {
		binding.Spec.Steps = append(binding.Spec.Steps, *encoder)
	}

	return nil
}

func configureDecoder(cc ConnectorConfiguration) (*camelv1alpha1.Endpoint, error) {
	var err error
	var step camelv1alpha1.Endpoint

	switch cc.DataShape.Consumes.Format {
	case "":
		break
	case ContentTypeJSON:
		step, err = endpoints.NewKameletBuilder("cos-decoder-json-action").Build()
	case ContentTypeAvroBinary:
		step, err = endpoints.NewKameletBuilder("cos-decoder-avro-action").Build()
	case ContentTypeText:
	case ContentTypeBinary:
		break
	default:
		err = fmt.Errorf("unsupported format %s", cc.DataShape.Consumes.Format)
	}

	if err != nil {
		return nil, errors.Wrap(err, "error creating sink")
	}

	if step.Ref != nil || step.URI != nil {
		return &step, nil
	}

	return nil, nil
}

func configureEncoder(cc ConnectorConfiguration) (*camelv1alpha1.Endpoint, error) {
	var err error
	var step camelv1alpha1.Endpoint

	switch cc.DataShape.Produces.Format {
	case "":
		break
	case ContentTypeJSON:
		step, err = endpoints.NewKameletBuilder("cos-encoder-json-action").Build()
	case ContentTypeAvroBinary:
		step, err = endpoints.NewKameletBuilder("cos-encoder-avro-action").Build()
	case ContentTypeText:
		step, err = endpoints.NewKameletBuilder("cos-encoder-string-action").Build()
	case ContentTypeBinary:
		step, err = endpoints.NewKameletBuilder("cos-encoder-bytearray-action").Build()
	default:
		err = fmt.Errorf("unsupported format %s", cc.DataShape.Produces.Format)
	}

	if err != nil {
		return nil, errors.Wrap(err, "error creating sink")
	}
	if step.Ref != nil || step.URI != nil {
		return &step, nil
	}

	return nil, nil
}
