package command

import (
	"context"
	"errors"
	"strings"

	"github.com/kaptinlin/jsonschema"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/logger"
)

type ValidateArgs struct {
	SchemaPath string
}

func (l *Lefthook) Validate(_ctx context.Context, args ValidateArgs) error {
	loader := config.NewLoader(l.repo, l.logger)
	main, secondary, err := loader.LoadKoanf()
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
		l.logValidationErrors(0, *details)

		return errors.New("validation failed for main config")
	}

	result = schema.Validate(secondary.Raw())
	if !result.IsValid() {
		details := result.ToList()
		l.logValidationErrors(0, *details)

		return errors.New("validation failed for secondary config")
	}

	l.logger.Info("All good")
	return nil
}

func (l *Lefthook) logValidationErrors(indent int, details jsonschema.List) {
	if details.Valid {
		return
	}

	if len(details.InstanceLocation) > 0 {
		l.logDetail(indent, details)

		indent += 2
	}

	for _, d := range details.Details {
		l.logValidationErrors(indent, d)
	}
}

func (l *Lefthook) logDetail(indent int, details jsonschema.List) {
	var errors []string
	if len(details.Errors) > 0 {
		for _, err := range details.Errors {
			errors = append(errors, err)
		}
	}

	option := strings.Repeat(" ", indent) + strings.TrimLeft(details.InstanceLocation, "/") + ":"

	if len(errors) == 0 {
		option = l.logger.Paint(logger.ColorGray, option)
	} else {
		option = l.logger.Paint(logger.ColorYellow, option)
	}

	if len(details.Details) > 0 {
		l.logger.Info(option)
	} else {
		l.logger.Info(option, l.logger.Paint(logger.ColorRed, strings.Join(errors, ",")))
	}
}
