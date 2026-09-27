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

func (r *Runtime) Install(pkg *model.AppPackage, signer ed25519.PublicKey, signature []byte) error {
	if pkg == nil || pkg.ID == "" || pkg.EntryPoint == "" || len(signer) != ed25519.PublicKeySize || !ed25519.Verify(signer, PackageDigest(pkg), signature) {
		return ErrInvalidPackageSignature
	}
	if old := r.packages[pkg.ID]; old != nil && pkg.Version <= old.Version {
		return ErrPackageDowngrade
	}
	if old := r.signers[pkg.ID]; old != nil && string(old) != string(signer) {
		return ErrPackageSignerChanged
	}
	copy := *pkg
	copy.Assets = copyAssets(pkg.Assets)
	r.packages[pkg.ID] = &copy
	r.signers[pkg.ID] = append([]byte(nil), signer...)
	return nil
}

func PackageDigest(pkg *model.AppPackage) []byte {
	hash := sha256.New()
	writePackageString(hash, pkg.ID)
	var version [8]byte
	binary.BigEndian.PutUint64(version[:], pkg.Version)
	hash.Write(version[:])
	writePackageString(hash, pkg.EntryPoint)
	keys := make([]string, 0, len(pkg.Assets))
	for name := range pkg.Assets {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		writePackageString(hash, name)
		asset := sha256.Sum256(pkg.Assets[name])
		hash.Write(asset[:])
	}
	return hash.Sum(nil)
}

func writePackageString(hash interface{ Write([]byte) (int, error) }, value string) {
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
