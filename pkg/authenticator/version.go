package authenticator

import "fmt"

// Version field is a SemVer that should indicate the baked-in version
// of the authn-k8s-client
var Version = "0.22.0"

// TagSuffix field denotes the specific build type for the client. It may
// be replaced by compile-time variables if needed to provide the git
// commit information in the final binary.
// In fixed versions, we don't want the tag to be present
var TagSuffix = "-dev"

// FullVersionName is the user-visible aggregation of version and tag
// of this codebase
var FullVersionName = fmt.Sprintf("%s%s", Version, TagSuffix)
