package model

type AppPackage struct {
	ID, EntryPoint string
	Version        uint64
	Assets         map[string][]byte
}
