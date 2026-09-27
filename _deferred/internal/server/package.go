package server

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sort"

	"mmos/internal/common/model"
)

var (
	ErrInvalidPackageSignature = errors.New("invalid package signature")
	ErrPackageDowngrade        = errors.New("package version must increase")
	ErrPackageSignerChanged    = errors.New("package signer changed")
)

// Install verifies a signed package before making it launchable. Packages
// supplied to NewRuntime are trusted provisioning inputs; all later changes
// go through this boundary.
func (r *Runtime) Install(pkg *model.AppPackage, signer ed25519.PublicKey, signature []byte) error {
	if pkg == nil || pkg.ID == "" || pkg.EntryPoint == "" || len(signer) != ed25519.PublicKeySize ||
		!ed25519.Verify(signer, packageDigest(pkg), signature) {
		return ErrInvalidPackageSignature
	}
	if installed := r.packages[pkg.ID]; installed != nil && pkg.Version <= installed.Version {
		return ErrPackageDowngrade
	}
	if previous := r.signers[pkg.ID]; previous != nil && string(previous) != string(signer) {
		return ErrPackageSignerChanged
	}
	copy := *pkg
	copy.Assets = copyAssets(pkg.Assets)
	r.packages[pkg.ID] = &copy
	r.signers[pkg.ID] = append([]byte(nil), signer...)
	return nil
}

// PackageDigest returns the deterministic bytes an installer signs. It is
// exported for packagers but verification stays server-owned.
func PackageDigest(pkg *model.AppPackage) []byte { return packageDigest(pkg) }

func packageDigest(pkg *model.AppPackage) []byte {
	hash := sha256.New()
	writeString(hash, pkg.ID)
	var version [8]byte
	binary.BigEndian.PutUint64(version[:], pkg.Version)
	hash.Write(version[:])
	writeString(hash, pkg.EntryPoint)
	keys := make([]string, 0, len(pkg.Assets))
	for name := range pkg.Assets {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		writeString(hash, name)
		asset := sha256.Sum256(pkg.Assets[name])
		hash.Write(asset[:])
	}
	return hash.Sum(nil)
}

func writeString(hash interface{ Write([]byte) (int, error) }, value string) {
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(value)))
	hash.Write(size[:])
	hash.Write([]byte(value))
}

func copyAssets(assets map[string][]byte) map[string][]byte {
	copy := make(map[string][]byte, len(assets))
	for name, asset := range assets {
		copy[name] = append([]byte(nil), asset...)
	}
	return copy
}
