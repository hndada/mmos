// Package protocol defines messages exchanged between an application process
// and the system runtime. The messages contain data only, so an in-process
// caller and a future IPC transport use the same contract.
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

// Launcher is the application-side view of the system launch endpoint.
//
// An IPC client can implement Launcher without exposing server state or
// credentials to the application process.
type Launcher interface {
	Launch(LaunchRequest) LaunchReply
}
