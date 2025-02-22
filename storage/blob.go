package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

type BlobStorage struct {
	containerClient *container.Client
	sharedKey       *azblob.SharedKeyCredential
	containerName   string
	accountName     string
}

func (s *BlobStorage) Upload(filename string, content io.Reader) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	blobClient := s.containerClient.NewBlockBlobClient(filename)
	_, err := blobClient.UploadStream(ctx, content, &azblob.UploadStreamOptions{
		Concurrency: 8,
		BlockSize:   8 * 1024 * 1024,
	})
	return err
}

func (s *BlobStorage) DownloadURL(filename string) string {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, filename)
}

func (s *BlobStorage) Delete(filename string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	blobClient := s.containerClient.NewBlockBlobClient(filename)
	if _, err := blobClient.Delete(ctx, nil); err != nil {
		return err
	}

	return nil
}

func NewBlobStorage(accountName string, accountKey string, containerName string) (*BlobStorage, error) {
	sharedKey, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}

	client, err := azblob.NewClientWithSharedKeyCredential(
		fmt.Sprintf("https://%s.blob.core.windows.net", accountName),
		sharedKey,
		nil,
	)
	if err != nil {
		return nil, err
	}
	containerClient := client.ServiceClient().NewContainerClient(containerName)

	return &BlobStorage{
		containerClient: containerClient,
		sharedKey:       sharedKey,
		containerName:   containerName,
		accountName:     accountName,
	}, nil
}
