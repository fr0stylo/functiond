package runner

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/mount"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/opencontainers/image-spec/identity"
	"golang.org/x/sys/unix"
)

type Snapshot struct {
	client      *containerd.Client
	snapshotter string
}

func NewSnapshot(client *containerd.Client) *Snapshot {
	return &Snapshot{client, defaults.DefaultSnapshotter}
}

func (r *Snapshot) CreateSnapshot(ctx context.Context, imageName string, name string, version string, zipPath string) (string, error) {
	image, err := r.client.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		log.Fatalf("failed to pull image: %v", err)
		return "", err
	}

	snapshotService := r.client.SnapshotService(r.snapshotter)
	digests, err := image.RootFS(ctx)

	imagefsid := identity.ChainID(digests).String()
	mounts, err := snapshotService.Prepare(ctx, name, imagefsid)
	if err != nil {
		log.Fatalf("failed to prepare writable snapshot: %v", err)
		return "", err
	}

	if err := unzipToSnapshot(mounts, zipPath); err != nil {
		log.Fatalf("failed to copy file to snapshot: %v", err)
	}
	commitName := fmt.Sprintf("%s-%s", name, version)
	if err := snapshotService.Commit(ctx, commitName, name); err != nil {
		return "", err
	}

	return commitName, nil
}

func unzipToSnapshot(mounts []mount.Mount, source string) error {
	if err := mount.All(mounts, "/mnt"); err != nil {
		return fmt.Errorf("failed to mount snapshot: %v", err)
	}
	defer mount.Unmount("/mnt", unix.MNT_DETACH)
	path := "/opt/function"
	if err := os.MkdirAll("/mnt"+path, 700); err != nil {
		return err
	}
	reader, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join("/mnt"+path, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
