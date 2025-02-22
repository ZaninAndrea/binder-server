package rest

import (
	"bytes"
	"encoding/base64"
	"strings"

	"github.com/ZaninAndrea/binder-server/storage"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

// ReplaceBase64ImagesWithFileLinks replaces base64 encoded images in the given HTML content with file links
func ReplaceBase64ImagesWithFileLinks(content string, storage *storage.BlobStorage) (string, error) {
	doc, _ := html.Parse(strings.NewReader(content))

	var crawlNode func(*html.Node) error
	crawlNode = func(node *html.Node) error {
		if node.Type == html.ElementNode && node.Data == "img" {
			// Check if the image is base64 encoded
			for i, attr := range node.Attr {
				// Upload the image to the server and replace the src attribute with the file link
				if attr.Key == "src" && strings.HasPrefix(attr.Val, "data:image/") {
					// Extract the image type from the src attribute
					const EXTENSION_START_POS = len("data:image/")
					imageType := strings.Split(attr.Val, ";")[0][EXTENSION_START_POS:]

					// Upload the image to the server
					imageContent := strings.SplitN(attr.Val, "base64", 2)[1][1:]
					ID := uuid.NewString() + "." + imageType
					data, err := base64.StdEncoding.DecodeString(imageContent)
					if err != nil {
						return err
					}

					err = storage.Upload(ID, bytes.NewReader(data))
					if err != nil {
						return err
					}

					// Replace the src attribute with the file link
					node.Attr[i].Val = storage.DownloadURL(ID)
					node.Attr = append(node.Attr, html.Attribute{Key: "az-blob-id", Val: ID})
				}
			}
			return nil
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := crawlNode(child); err != nil {
				return err
			}
		}

		return nil
	}
	if err := crawlNode(doc); err != nil {
		return "", err
	}

	// Render the modified HTML
	var b strings.Builder
	err := html.Render(&b, doc)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// ListImageIDs returns a list of image IDs from the given HTML content
func ListImageIDs(content string) []string {
	doc, _ := html.Parse(strings.NewReader(content))

	var imageIDs []string
	var crawlNode func(*html.Node)
	crawlNode = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "img" {
			for _, attr := range node.Attr {
				if attr.Key == "az-blob-id" {
					imageIDs = append(imageIDs, attr.Val)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawlNode(child)
		}
	}
	crawlNode(doc)

	return imageIDs
}

// Diff returns the added and removed image IDs between the two given HTML contents
func Diff(oldContent, newContent string) (added, removed []string) {
	oldImageIDs := ListImageIDs(oldContent)
	newImageIDs := ListImageIDs(newContent)

	// Find the added image IDs
	added = make([]string, 0)
	for _, id := range newImageIDs {
		if !slices.Contains(oldImageIDs, id) {
			added = append(added, id)
		}
	}

	// Find the removed image IDs
	removed = make([]string, 0)
	for _, id := range oldImageIDs {
		if !slices.Contains(newImageIDs, id) {
			removed = append(removed, id)
		}
	}

	return added, removed
}
