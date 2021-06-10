package api

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/os-pc/gocloudlb"
	"github.com/os-pc/gocloudlb/loadbalancers"
	"github.com/os-pc/gocloudlb/nodes"
	"github.com/os-pc/gocloudlb/virtualips"

	//"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	//"github.com/gophercloud/gophercloud/pagination"
	"github.com/gophercloud/utils/openstack/clientconfig"

	"github.com/sulochan/kaas/models"
)

func GetOpenstackProvider(authOpts models.AuthOpts) (*gophercloud.ProviderClient, error) {
	opts := &clientconfig.ClientOpts{}

	if authOpts.Type == "Token" {
		fmt.Println("Token passed for auth")
		opts = &clientconfig.ClientOpts{
			AuthType: clientconfig.AuthV2Token,
			AuthInfo: &clientconfig.AuthInfo{
				AuthURL:     "https://lon.identity.api.rackspacecloud.com/v2.0/",
				Username:    authOpts.Username,
				Token:       authOpts.Token,
				ProjectName: authOpts.ProjectId,
				DomainName:  authOpts.ProjectId,
			},
		}
	} else if authOpts.Type == "Password" {
		fmt.Println("Password passed for auth")
		opts = &clientconfig.ClientOpts{
			AuthInfo: &clientconfig.AuthInfo{
				AuthURL:     "https://lon.identity.api.rackspacecloud.com/v2.0/",
				Username:    authOpts.Username,
				Password:    authOpts.Password,
				ProjectName: authOpts.ProjectId,
				DomainName:  authOpts.ProjectId,
			},
		}
	} else {
		fmt.Println("No auth options passed in the headers.")
	}

	provider, err := clientconfig.AuthenticatedClient(opts)
	if err != nil {
		fmt.Println(err)
		return provider, err
	}

	return provider, err
}

// GetComputeServcie - get openstack compute cli
func GetComputeServcie(authOpts models.AuthOpts) (*gophercloud.ServiceClient, error) {
	provider, err := GetOpenstackProvider(authOpts)
	if err != nil {
		fmt.Println(err)
		return &gophercloud.ServiceClient{}, err
	}

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		//Region: "DFW",
		Region: "LON",
	})

	return client, err
}

func CreateVM(clusterName string, serverType string, count int, authOpts models.AuthOpts) (*models.Node, error) {
	client, err := GetComputeServcie(authOpts)
	if err != nil {
		fmt.Println("failed to get compute client ", err)
		return &models.Node{}, err
	}

	content, err := ioutil.ReadFile("golangcode.txt")
	if err != nil {
		log.Fatal(err)
	}
	serverData := string(content)
	fmt.Println(serverData)

	configDrive := true
	servername := fmt.Sprintf("k8s-%s-%s-%v", clusterName, serverType, count)
	server, err := servers.Create(client, servers.CreateOpts{
		Name: servername,
		//FlavorRef:     "performance1-1",
		FlavorRef: "5",
		//ImageRef:  "e2ddb657-30ce-4c40-9b53-e884a65d1df0",
		ImageRef:    "e83e244d-af6a-4b68-a4cc-a425897021af",
		Metadata:    map[string]string{"k8saas": "true", "cluster": clusterName},
		UserData:    []byte(serverData),
		ConfigDrive: &configDrive,
		//ServiceClient: client,
	}).Extract()
	if err != nil {
		fmt.Println("Unable to create server: %s", err)
		return &models.Node{}, err
	}

	serverNode := models.Node{Name: servername, UUID: server.ID, Password: server.AdminPass}
	fmt.Println("Returning serverNode -> ", serverNode)
	return &serverNode, nil
}

// DeleteVM - delete cs vm.
func DeleteVM(uuid string, authOpts models.AuthOpts) error {
	client, err := GetComputeServcie(authOpts)
	if err != nil {
		fmt.Println("failed to get compute client ", err)
		return err
	}
	result := servers.Delete(client, uuid)
	fmt.Println(result)
	return nil
}

// GetLbaasService - get Rackspace lbaas service
func GetLbaasService(authOpts models.AuthOpts) (*gophercloud.ServiceClient, error) {
	provider, err := GetOpenstackProvider(authOpts)
	if err != nil {
		return &gophercloud.ServiceClient{}, err
	}

	client, err := gocloudlb.NewLB(provider, gophercloud.EndpointOpts{
		Region: "LON",
	})

	return client, err

}

func createLoadbalancer(lb loadbalancers.LoadBalancer, authOpts models.AuthOpts) (*loadbalancers.LoadBalancer, error) {
	lbaasClient, err := GetLbaasService(authOpts)
	if err != nil {
		fmt.Println("Making client: ", err)
	}

	viptype := virtualips.CreateOpts{Type: "PUBLIC"}

	opts := loadbalancers.CreateOpts{
		Name:       lb.Name,
		Port:       lb.Port,
		Protocol:   lb.Protocol,
		VirtualIps: []virtualips.CreateOpts{viptype},
		Nodes:      []nodes.CreateOpts{},
		//            Nodes: []nodes.Node{
		//               nodes.Node{Address: "10.1.1.1", Port: 80, Condition: nodes.ENABLED},
		//            },
	}

	fmt.Println(opts)
	lbout, err := loadbalancers.Create(lbaasClient, opts).Extract()
	log.Println(lbout)
	return lbout, err
}

func deleteLoadbalancer(raxlb *loadbalancers.LoadBalancer, authOpts models.AuthOpts) error {
	lbaasClient, err := GetLbaasService(authOpts)
	if err != nil {
		fmt.Println("Making client: ", err)
	}

	result := loadbalancers.Delete(lbaasClient, raxlb.ID)
	//if err != nil {
	//	fmt.Println("Error deleting LB.")
	//}
	fmt.Println(result)
	return err
}

func getLoadbalancer(raxlb *loadbalancers.LoadBalancer, authOpts models.AuthOpts) *loadbalancers.LoadBalancer {
	lbaasClient, err := GetLbaasService(authOpts)
	if err != nil {
		fmt.Println("Making client: ", err)
	}

	lb := loadbalancers.Get(lbaasClient, raxlb.ID)
	newraxlb, err := lb.Extract()
	if err != nil {
		fmt.Println("Error extracting lb: ", err)
	}
	return newraxlb
}

func getAllLoadbalancers(authOpts models.AuthOpts) []loadbalancers.LoadBalancer {
	lbassClient, err := GetLbaasService(authOpts)
	if err != nil {
		fmt.Println("Making client: ", err)
	}

	lbpager := loadbalancers.List(lbassClient, nil)
	if err != nil {
		fmt.Println("All pages: ", err)
	}

	lblist := []loadbalancers.LoadBalancer{}

	err = lbpager.EachPage(func(page pagination.Page) (bool, error) {
		lbList, err := loadbalancers.ExtractLoadBalancers(page)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		for _, s := range lbList {
			lblist = append(lblist, s)
		}
		return true, nil
	})

	return lblist
}

func attachNodesToLoadbalancer(raxlb *loadbalancers.LoadBalancer, lbnodes []string, authOpts models.AuthOpts) {
	lbaasClient, err := GetLbaasService(authOpts)
	if err != nil {
		fmt.Println("Making client: ", err)
	}

	opts := []nodes.CreateOpts{}
	for _, i := range lbnodes {
		n := nodes.CreateOpts{Address: i, Port: raxlb.Port, Condition: "ENABLED"}
		opts = append(opts, n)
	}

	nodeList := nodes.Create(lbaasClient, raxlb.ID, opts)
	fmt.Println("Created Node list: ", nodeList)
}
