package lefthook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kaptinlin/jsonschema"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/log"
)

const schemaUrl = "https://raw.githubusercontent.com/evilmartians/lefthook/master/schema.json"

func Validate(opts *Options) error {
	lefthook, err := initialize(opts)
	if err != nil {
		return fmt.Errorf("couldn't initialize lefthook: %w", err)
	}

	main, secondary, err := config.LoadKoanf(lefthook.Fs, lefthook.repo)
	if err != nil {
		return err
	}

	compiler := jsonschema.NewCompiler()

	schema, err := compiler.GetSchema(schemaUrl)
	if err != nil {
		return err
	}

	result := schema.Validate(main.Raw())
	if !result.IsValid() {
		details := result.ToList()
		logValidationErrors(0, *details)

		return errors.New("Validation failed for main config")
	}

	result = schema.Validate(secondary.Raw())
	if !result.IsValid() {
		details := result.ToList()
		logValidationErrors(0, *details)

		return errors.New("Validation failed for secondary config")
	}

	log.Info("All good")
	return nil
}

func logValidationErrors(indent int, details jsonschema.List) {
	if details.Valid {
		return
	}

	if len(details.InstanceLocation) > 0 {
		logDetail(indent, details)

		indent += 2
	}

	for _, d := range details.Details {
		logValidationErrors(indent, d)
	}
}

func logDetail(indent int, details jsonschema.List) {
	var errors []string
	if len(details.Errors) > 0 {
		for _, err := range details.Errors {
			errors = append(errors, err)
		}
	}

	option := strings.Repeat(" ", indent) + strings.TrimLeft(details.InstanceLocation, "/") + ":"

	if len(errors) == 0 {
		option = log.Gray(option)
	} else {
		option = log.Yellow(option)
	}

	if len(details.Details) > 0 {
		log.Info(option)
	} else {
		log.Info(option, log.Red(strings.Join(errors, ",")))
	}
}
