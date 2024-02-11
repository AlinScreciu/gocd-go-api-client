package client

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/packages"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

func (c *Client) GetAllPackages() (*types.AllPackages, error) {
	return packages.GetAllPackages(c.client)
}

func (c *Client) CreatePackage(pkg *types.Package) (*types.Package, error) {
	return packages.CreatePackage(c.client, pkg)
}

func (c *Client) GetPackage(packageId string) (*types.Package, error) {
	return packages.GetPackage(c.client, packageId)
}

func (c *Client) GetPackageWithETag(packageId string) (*types.Package, string, error) {
	return packages.GetPackageWithETag(c.client, packageId)
}

func (c *Client) UpdatePackage(pkg *types.Package, eTag string) (*types.Package, error) {
	return packages.UpdatePackage(c.client, pkg, eTag)
}

func (c *Client) DeletePackage(packageId string) (string, error) {
	return packages.DeletePackage(c.client, packageId)
}
