package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"regexp"
	"strings"
	"text/template"
)

var (
	cmd      = flag.String("cmd", "", "Name of the subcommand")
	out      = flag.String("out", "", "Output file")
	pkg      = flag.String("pkg", "", "Name of the package")
	synopsis = flag.String("synopsis", "", "Synopsis of the subcommand")
	usage    = flag.String("usage", "", "Usage example of the subcommand")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `generate-subcommand: A code generator for subcommands.

The resulting file will contain a type which satifies the subcommands.Command
interface. See https://godoc.org/github.com/google/subcommands.
The command accepts all parameters in the form of flags; however, flags
that have not been specified will be prompted for. Not all inputs are
required.

Usage: generate-subcommand [flags]`)
		fmt.Fprintln(os.Stderr, "\nFlags:")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)

	// If the subcommand name was not specified via flags, continually
	// prompt for confirmation until one is provided.
	if *cmd == "" {
		for *cmd == "" {
			fmt.Print("Enter subcommand name: ")
			scanner.Scan()
			*cmd = strings.TrimSpace(wordRegex.FindString(scanner.Text()))
		}
	}

	// If the output filename was not specified via flags, prompt for
	// confirmation. Provide a default option that matches the name of the
	// command.
	if *out == "" {
		for *out == "" {
			fmt.Printf("Enter out file [%s.go]: ", *cmd)
			scanner.Scan()
			*out = func() string {
				if scanner.Text() == "" {
					return *cmd + ".go"
				}
				if in := strings.TrimSpace(fileRegex.FindString(scanner.Text())); in != "" {
					if !strings.HasSuffix(in, ".go") {
						in += ".go"
					}
					return in
				}
				return ""
			}()
		}
	}

	// If the provided output file already exists, prompt for confirmation.
	// Exit the program if the user declines.
	if _, err := os.Stat(*out); err == nil {
		for ok := false; !ok; {
			fmt.Printf("File %q exists, overwrite? [y/N]: ", *out)
			scanner.Scan()
			switch strings.ToLower(scanner.Text()) {
			case "y", "ye", "yes":
				ok = true
			case "n", "no", "":
				os.Exit(0)
			default:
			}
		}
	}

	// If a package was not specified via flags, assume that this command
	// will be part of the main package.
	if *pkg == "" {
		for *pkg == "" {
			fmt.Printf("Enter package name [%s]: ", *cmd)
			scanner.Scan()
			*pkg = func() string {
				if scanner.Text() == "" {
					return *cmd
				}
				return strings.TrimSpace(wordRegex.FindString(scanner.Text()))
			}()
		}
	}

	// If the package is specified as something other than package main,
	// assume that the user will want command to be exported.
	if *pkg != "main" {
		*cmd = strings.Title(*cmd)
	}

	if *synopsis == "" {
		fmt.Print("Enter synopsis? [y/N]: ")
		scanner.Scan()
		switch strings.ToLower(scanner.Text()) {
		case "y", "ye", "yes":
			scanner.Scan()
			*synopsis = scanner.Text()
		}
	}

	if *usage == "" {
		fmt.Print("Enter usage? [y/N]: ")
		scanner.Scan()
		switch strings.ToLower(scanner.Text()) {
		case "y", "ye", "yes":
			fmt.Println("^D to end.")
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			*usage = strings.Join(lines, "\n")
		}
	}

	// The username is used to assign a name to TODOs in the generated file.
	username := "somebody"
	if usr, err := user.Current(); err == nil {
		// It's not essential that a user is retrieved, so the error is
		// not handled.
		username = usr.Username
	}

	var buf bytes.Buffer

	err := tmpl.Execute(&buf, tmplParams{
		*cmd, *pkg, *synopsis, *usage, username,
	})
	if err != nil {
		// A failure executing the template signifies an unrecoverable
		// problem with the program, and an incorrect file should not be
		// generated.
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(*out, buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

// wordRegex is a regular expression used to validate user input, where the
// wanted input is a single word.
// The first character must be alphabetical as Go identifiers cannot start with
// a number.
// The regex accepts leading and trailing spaces as a convinience for the user,
var wordRegex = regexp.MustCompile(`^\s*[A-Za-z][A-Za-z0-9]+`)

// fileRegex is a regular expression used to validate user input, where the
// wanted input is a single file name (.go suffix is optional).
// The first character must be alphabetical.
// The regex accepts leading and trailing spaces as a convinience for the user,
var fileRegex = regexp.MustCompile(`^\s*[A-Za-z][A-Za-z0-9]+(.go)?`)

type tmplParams struct {
	Cmd, Pkg, Synopsis, Usage, User string
}

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"First": func(s string) string { return string(s[0]) },
	"Lower": strings.ToLower,
}).Parse(`package {{ .Pkg }}

import (
	"context"
	"flag"

	"github.com/google/subcommands"
)

type {{ .Cmd }} struct{}

func (*{{ .Cmd }}) Name() string {
	return "{{ .Cmd | Lower }}"
}

func (*{{ .Cmd }}) Synopsis() string {
	return "{{ .Synopsis }}"
}

func (*{{ .Cmd }}) Usage() string {
	return ` + "`{{ .Usage }}`" + `
}

func ({{ .Cmd | Lower | First }} *{{ .Cmd }}) SetFlags(f *flag.FlagSet) {}

func ({{ .Cmd | Lower | First }} *{{ .Cmd }}) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// TODO({{ .User }})

	return subcommands.ExitUsageError
}
`))
