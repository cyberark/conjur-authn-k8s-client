package authenticator

import "fmt"

// Version field is a SemVer that should indicate the baked-in version
// of the authn-k8s-client
var Version = "0.16.1"

// Tag field denotes the specific build type for the client.
var Tag = "dev"

// FullVersionName is the user-visible aggregation of version and tag
// of this codebase
var FullVersionName = fmt.Sprintf("%s-%s", Version, Tag)