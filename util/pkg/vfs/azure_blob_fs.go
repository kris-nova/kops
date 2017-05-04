package vfs

import (
	"path"
	"fmt"
	"encoding/base64"
	"github.com/Azure/azure-sdk-for-go/storage"
	"io/ioutil"
)

// https://kopsdevel.blob.core.windows.net
type AzureBlobPath struct {
	azureBlobContext *AzureBlobContext
	container        string
	key              string
}

func newAzureBlobPath(azureBlobCtx *AzureBlobContext, container string, key string) (*AzureBlobPath) {
	return &AzureBlobPath{
		container:        container,
		key:              key,
		azureBlobContext: azureBlobCtx,
	}
}

func (a *AzureBlobPath) Join(relativePath ...string) Path {
	args := []string{a.key}
	args = append(args, relativePath...)
	joined := path.Join(args...)
	return &AzureBlobPath{
		azureBlobContext: a.azureBlobContext,
		container:        a.container,
		key:              joined,
	}
}

func (a *AzureBlobPath) ReadFile() ([]byte, error) {
	client, err := a.azureBlobContext.getClient()
	var retBytes []byte
	if err != nil {
		return retBytes, fmt.Errorf("unable to get azure storage blob client for file %s: %v", a.key, err)
	}
	blobClient := client.GetBlobService()

	readCloser, err := blobClient.GetBlob(a.container, a.key)
	if err != nil {
		fmt.Println("----------")
		fmt.Println(a.container)
		fmt.Println(a.key)
		fmt.Println("----------")
		return retBytes, fmt.Errorf("unable to get blob: %v", err)
	}
	retBytes, err = ioutil.ReadAll(readCloser)
	defer readCloser.Close()
	if err != nil {
		return retBytes, fmt.Errorf("unable to read bytes: %v", err)
	}
	return retBytes, nil
}

func (a *AzureBlobPath) WriteFile(data []byte) error {
	client, err := a.azureBlobContext.getClient()
	if err != nil {
		return fmt.Errorf("unable to get azure storage blob client: %v", err)
	}
	blobClient := client.GetBlobService()

	// Create a block for the data
	blockID := base64.StdEncoding.EncodeToString([]byte(a.key))
	blobClient.PutBlock(a.container, a.key, blockID, data)

	// Get block list
	blocksList, err := blobClient.GetBlockList(a.container, a.key, storage.BlockListTypeUncommitted)
	if err != nil {
		return err
	}

	// Build uncommitted blocks list
	uncommittedBlocksList := make([]storage.Block, len(blocksList.UncommittedBlocks))
	for i := range blocksList.UncommittedBlocks {
		uncommittedBlocksList[i].ID = blocksList.UncommittedBlocks[i].Name
		uncommittedBlocksList[i].Status = storage.BlockStatusUncommitted
	}

	// Write the blocks to the blob
	blobClient.PutBlockList(a.container, a.key, uncommittedBlocksList)
	return nil
}

func (a *AzureBlobPath) CreateFile(data []byte) error {
	client, err := a.azureBlobContext.getClient()
	if err != nil {
		return fmt.Errorf("unable to get azure storage blob client: %v", err)
	}
	blobClient := client.GetBlobService()
	err = blobClient.CreateBlockBlob(a.container, a.key)
	if err != nil {
		return fmt.Errorf("unable to create block blob: %v", err)
	}
	return nil
}

func (a *AzureBlobPath) Base() string {
	return path.Base(a.key)
}
func (a *AzureBlobPath) Path() string {
	return "https://" + a.container + ".blob.core.windows.net/" + a.container + "/" + a.key
}

func (a *AzureBlobPath) Remove() error {
	client, err := a.azureBlobContext.getClient()
	if err != nil {
		return fmt.Errorf("unable to get azure storage blob client: %v", err)
	}
	blobClient := client.GetBlobService()
	if _, err = blobClient.DeleteBlobIfExists(a.container, a.key, nil); err != nil {
		return fmt.Errorf("unable to remove blob: %v", err)
	}
	return nil
}

func (a *AzureBlobPath) ReadDir() ([]Path, error) {
	var paths []Path
	client, err := a.azureBlobContext.getClient()
	if err != nil {
		return paths, fmt.Errorf("unable to get azure storage blob client: %v", err)
	}
	blobClient := client.GetBlobService()
	list, err := blobClient.GetBlockList(a.container, a.key, storage.BlockListTypeAll)
	if err != nil {
		return paths, fmt.Errorf("unable to list block list: %v", err)
	}
	for _, l := range list.CommittedBlocks {
		p := &AzureBlobPath{
			container: a.container,
			key: l.Name,
		}
		paths = append(paths, p)
	}
	return paths, nil
}
func (a *AzureBlobPath) ReadTree() ([]Path, error) {
	var paths []Path
	//paths = append(paths, AzureBlobPath{})
	return paths, nil
}
