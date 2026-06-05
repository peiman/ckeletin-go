// cmd/catalog.go
//
// ckeletin:allow-custom-command — a framework introspection command whose
// cobra-specific walk cannot live in internal/ (cobra is the commands layer
// only, CKSPEC-ARCH-003) and is not config-registry-driven, so it does not
// follow the NewCommand/metadata pattern. Same exemption as cmd/config.go.
//
// `catalog` command — emits the machine-readable command catalog
// (CKSPEC-AGENT-006), derived from the SAME cobra command tree the parser uses
// (RootCmd). Because the catalog and the parser read one tree, the catalog
// cannot drift from the actual command set: anti-drift is structural, not
// tested-in.
//
// The Catalog *types* are the framework's shared schema (.ckeletin/pkg/catalog);
// this file is the cobra-specific walk — the only implementation-dependent part.
// cobra is confined to the commands layer (CKSPEC-ARCH-003), so the walk lives
// here, not in business/infrastructure, mirroring ckeletin-rust where the clap
// walk lives in the cli crate.

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/catalog"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Emit the machine-readable command catalog",
	Long: `Emit the CLI's own command surface as structured data (CKSPEC-AGENT-006).

The catalog enumerates every command, its subcommands, and flags — which are
required and which take a value — derived from the actual command definitions so
it cannot drift from the real surface. With --output json it is the data of a
success envelope, letting an agent discover the full capability surface without
parsing human --help text.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE:         runCatalog,
}

func init() {
	MustAddToRoot(catalogCmd)
}

func runCatalog(cmd *cobra.Command, _ []string) error {
	cat := buildCatalog(cmd.Root())

	if output.IsJSONMode() {
		return output.RenderJSON(cmd.OutOrStdout(), output.JSONEnvelope{
			Status:  "success",
			Command: output.CommandName(),
			Data:    cat,
		})
	}

	_, err := fmt.Fprint(cmd.OutOrStdout(), cat.String())
	return err
}

// buildCatalog derives the catalog from a cobra command tree. Global (persistent)
// flags are collected once at the top level, not duplicated into each command.
func buildCatalog(root *cobra.Command) catalog.Catalog {
	return catalog.Catalog{
		Name:        root.Name(),
		Description: root.Short,
		GlobalFlags: catalogFlags(root.PersistentFlags()),
		Commands:    walkCommands(root.Commands()),
	}
}

func walkCommands(cmds []*cobra.Command) []catalog.Command {
	out := make([]catalog.Command, 0, len(cmds))
	for _, c := range cmds {
		if !includeInCatalog(c) {
			continue
		}
		out = append(out, catalog.Command{
			Name:        c.Name(),
			Description: c.Short,
			Flags:       catalogFlags(c.LocalNonPersistentFlags()),
			Commands:    walkCommands(c.Commands()),
		})
	}
	return out
}

// includeInCatalog drops cobra's auto-generated `help` command and any hidden or
// otherwise-unavailable command; the application's own surface stays — including
// the catalog command itself (self-referential) and the completion command.
func includeInCatalog(c *cobra.Command) bool {
	if c.Name() == "help" {
		return false
	}
	return c.IsAvailableCommand()
}

func catalogFlags(fs *pflag.FlagSet) []catalog.Flag {
	out := make([]catalog.Flag, 0)
	fs.VisitAll(func(f *pflag.Flag) {
		// --help / --version are universal cobra scaffolding, not part of the
		// application's surface; hidden flags MAY be excluded (R2).
		if f.Name == "help" || f.Name == "version" || f.Hidden {
			return
		}
		out = append(out, catalogFlag(f))
	})
	return out
}

func catalogFlag(f *pflag.Flag) catalog.Flag {
	// A boolean switch reports value type "bool"; everything else consumes a value.
	takesValue := f.Value.Type() != "bool"

	cf := catalog.Flag{
		Long:        f.Name,
		Required:    flagRequired(f),
		TakesValue:  takesValue,
		Short:       f.Shorthand,
		Description: f.Usage,
	}
	// Emit a default only for value flags with a real default — a switch's "false"
	// default is noise (mirrors the rust hygiene). possible_values is always
	// omitted: cobra cannot derive enumerated flag values structurally.
	if takesValue && f.DefValue != "" {
		cf.Default = f.DefValue
	}
	return cf
}

// flagRequired reports whether MarkFlagRequired was called on the flag. cobra
// stores required-ness as the BashCompOneRequiredFlag annotation, not a getter.
func flagRequired(f *pflag.Flag) bool {
	vals, ok := f.Annotations[cobra.BashCompOneRequiredFlag]
	return ok && len(vals) == 1 && vals[0] == "true"
}
