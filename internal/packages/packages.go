package packages

import (
	"github.com/AlinScreciu/gocd-go-api-client/internal/client"
	"github.com/AlinScreciu/gocd-go-api-client/internal/constants"
	"github.com/AlinScreciu/gocd-go-api-client/pkg/types"
)

const (
	endpoint = "/api/admin/packages"
)

func GetAllPackages(c *client.Client) (*types.AllPackages, error) {
	return client.Get[types.AllPackages](c, endpoint, constants.AcceptV2, "packages")
}

func GetPackage(c *client.Client, packageId string) (*types.Package, error) {
	return client.Get[types.Package](c, endpoint+"/"+packageId, constants.AcceptV2, "packages")
}

func GetPackageWithETag(c *client.Client, packageId string) (*types.Package, string, error) {
	return client.GetWithETag[types.Package](c, endpoint+"/"+packageId, constants.AcceptV2, "packages")
}

func CreatePackage(c *client.Client, pkg *types.Package) (*types.Package, error) {
	return client.Post[types.Package, types.Package](c, pkg, endpoint, constants.AcceptV2, "packages")
}

func UpdatePackage(c *client.Client, pkg *types.Package, eTag string) (*types.Package, error) {
	return client.Put[types.Package, types.Package](c, pkg, eTag, endpoint+"/"+pkg.Id, constants.AcceptV2, "packages")
}

func DeletePackage(c *client.Client, packageId string) (string, error) {
	return client.Delete(c, endpoint+"/"+packageId, constants.AcceptV2, "packages")
}
