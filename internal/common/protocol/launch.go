package protocol

// LaunchRequest asks the system to start or resume an installed package.
type LaunchRequest struct {
	PackageID string
}

// LaunchReply reports whether the system started or resumed the package.
type LaunchReply struct {
	Started bool
	Reused  bool
}
