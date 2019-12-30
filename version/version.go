package version

// Build information. Populated at build-time by the build script.
var (
	APIVersion      string
	OperatorVersion string
	CommitSHA       string
	Branch          string
	CommitTimestamp string
	GoVersion       string
)

// Info provides the iterable version information.
var Info = map[string]string{
	"apiVersion":      APIVersion,
	"operatorVersion": OperatorVersion,
	"commitSHA":       CommitSHA,
	"branch":          Branch,
	"commitTimestamp": CommitTimestamp,
	"goVersion":       GoVersion,
}

// VersionFormatStr is the format string to use for printing version
// information of the build.
var VersionFormatStr = `APIVersion       : %s
Operator Version : %s
Commit SHA-1     : %s
Commit Timestamp : %s
Branch           : %s
Go Version       : %s
`
