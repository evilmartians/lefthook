package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/invopop/jsonschema"

	"github.com/evilmartians/lefthook/internal/config"
)

//go:generate go run jsonschema.go
func main() {
	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true
	r.AdditionalFields = func(t reflect.Type) []reflect.StructField {
		if t == reflect.TypeOf(config.Config{}) {
			return reflect.VisibleFields(reflect.TypeOf(struct {
				Schema               string       `json:"$schema,omitempty"`
				PreCommit            *config.Hook `json:"pre-commit,omitempty"`
				ApplypatchMsg        *config.Hook `json:"applypatch-msg,omitempty"`
				PreApplypatch        *config.Hook `json:"pre-applypatch,omitempty"`
				PostApplypatch       *config.Hook `json:"post-applypatch,omitempty"`
				PreMergeCommit       *config.Hook `json:"pre-merge-commit,omitempty"`
				PrepareCommitMsg     *config.Hook `json:"prepare-commit-msg,omitempty"`
				CommitMsg            *config.Hook `json:"commit-msg,omitempty"`
				PostCommit           *config.Hook `json:"post-commit,omitempty"`
				PreRebase            *config.Hook `json:"pre-rebase,omitempty"`
				PostCheckout         *config.Hook `json:"post-checkout,omitempty"`
				PostMerge            *config.Hook `json:"post-merge,omitempty"`
				PrePush              *config.Hook `json:"pre-push,omitempty"`
				PreReceive           *config.Hook `json:"pre-receive,omitempty"`
				Update               *config.Hook `json:"update,omitempty"`
				ProcReceive          *config.Hook `json:"proc-receive,omitempty"`
				PostReceive          *config.Hook `json:"post-receive,omitempty"`
				PostUpdate           *config.Hook `json:"post-update,omitempty"`
				ReferenceTransaction *config.Hook `json:"reference-transaction,omitempty"`
				PushToCheckout       *config.Hook `json:"push-to-checkout,omitempty"`
				PreAutoGc            *config.Hook `json:"pre-auto-gc,omitempty"`
				PostRewrite          *config.Hook `json:"post-rewrite,omitempty"`
				SendemailValidate    *config.Hook `json:"sendemail-validate,omitempty"`
				FsmonitorWatchman    *config.Hook `json:"fsmonitor-watchman,omitempty"`
				P4Changelist         *config.Hook `json:"p4-changelist,omitempty"`
				P4PrepareChangelist  *config.Hook `json:"p4-prepare-changelist,omitempty"`
				P4PostChangelist     *config.Hook `json:"p4-post-changelist,omitempty"`
				P4PreSubmit          *config.Hook `json:"p4-pre-submit,omitempty"`
				PostIndexChange      *config.Hook `json:"post-index-change,omitempty"`
			}{}))
		}

		return []reflect.StructField{}
	}
	schema := r.Reflect(&config.Config{})
	schema.ID = "https://json.schemastore.org/lefthook.json"
	schema.Comments = "Last updated on " + time.Now().Format("2006.01.02") + "."
	dumped, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to generate json: %s", err)
		os.Exit(1)
	}

	_, _ = os.Stdout.Write(dumped)
}
