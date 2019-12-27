package version

// Build information. Populated at build-time by the build script.
var (
	Version   string
	Revision  string
	Branch    string
	BuildDate string
	GoVersion string
)

// Info provides the iterable version information.
var Info = map[string]string{
	"version":   Version,
	"revision":  Revision,
	"branch":    Branch,
	"buildDate": BuildDate,
	"goVersion": GoVersion,
}

var VersionFormatStr = `Version    : %s
Revision   : %s
Branch     : %s
Build-Date : %s
Go-Version : %s
`
