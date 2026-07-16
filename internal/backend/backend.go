// Package backend accesses files on an MTP device.
package backend

import "context"

// Device is a detected MTP device.
type Device struct {
	Bus       int
	Dev       int
	VendorID  uint16
	ProductID uint16
	Vendor    string
	Product   string
	Serial    string
}

// Object is a file or directory stored on a device.
type Object struct {
	ID        uint32
	ParentID  uint32
	StorageID uint32
	Name      string
	Path      string
	IsDir     bool
	Size      int64
}

// Backend accesses an MTP device. Every method takes a context so callers can
// cancel long-running transfers.
type Backend interface {
	// Detect returns the MTP devices currently attached.
	Detect(ctx context.Context) ([]Device, error)
	// List returns every object (files and folders) on the device.
	List(ctx context.Context) ([]Object, error)
	// Get copies the object with the given id to destPath on the local disk.
	Get(ctx context.Context, id uint32, destPath string) error
	// Put copies a local file to remotePath on the device and returns the new
	// object id (0 if the backend cannot report it).
	Put(ctx context.Context, localPath, remotePath string) (uint32, error)
	// Delete removes the object with the given id from the device.
	Delete(ctx context.Context, id uint32) error
}
