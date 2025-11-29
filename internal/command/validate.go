package command

import (
	"context"
	"errors"
	"strings"

	"github.com/kaptinlin/jsonschema"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/log"
)

type ValidateArgs struct {
	SchemaPath string
}

func (l *Lefthook) Validate(_ctx context.Context, args ValidateArgs) error {
	main, secondary, err := config.LoadKoanf(l.fs, l.repo)
	if err != nil {
		return err
	}

	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(config.JsonSchema)
	if err != nil {
		return err
	}

	result := schema.Validate(main.Raw())
	if !result.IsValid() {
		details := result.ToList()
		logValidationErrors(0, *details)

		return errors.New("validation failed for main config")
	}

	result = schema.Validate(secondary.Raw())
	if !result.IsValid() {
		details := result.ToList()
		logValidationErrors(0, *details)

		return errors.New("validation failed for secondary config")
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
