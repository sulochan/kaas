package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"net/http"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/os-pc/gocloudlb/loadbalancers"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	db "github.com/sulochan/kaas/db/mongodb"
	"github.com/sulochan/kaas/models"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// GetAllClusters - get k8s clusters
func GetCluster(w http.ResponseWriter, r *http.Request) {
	authOpts := context.Get(r, "authOpts").(models.AuthOpts)

	fmt.Println("Get clusters got called")
	public := models.Public{}
	names := []string{}
	data := []models.Node{}

	client, err := GetComputeServcie(authOpts)
	if err != nil {
		http.Error(w, "Error creating openstack client to get servers", 500)
		return
	}
	sopts := servers.ListOpts{Name: "k8s-*"}
	pager := servers.List(client, sopts)

	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		for _, s := range serverList {
			// "s" will be a servers.Server
			cluster, cluster_ok := s.Metadata["cluster"]
			roles, roles_ok := s.Metadata["roles"]

			_, ok := s.Metadata["k8saas"]
			if ok {
				node := models.Node{IP: s.AccessIPv4}
				if cluster_ok {
					if !stringInSlice(cluster, names) {
						names = append(names, cluster)
					}
					node.Cluster = cluster
				}
				if roles_ok {
					node.Roles = []string{roles}
				}
				data = append(data, node)
			}
		}
		return true, nil
	})

	public.Names = names
	public.Nodes = data

	publicData := map[string][]models.Node{}
	for _, j := range public.Names {
		for _, l := range public.Nodes {
			if l.Cluster == j {
				publicData[j] = append(publicData[j], l)
			}
		}
	}
	json.NewEncoder(w).Encode(publicData)
	return

}

// GetCluster - get available clusters for this account
func GetAllClusters(w http.ResponseWriter, r *http.Request) {
	projectid := context.Get(r, "projectid")

	clusters, err := db.GetAllClusters(projectid.(string))
	if err != nil {
		fmt.Println(err)
	}

	type resp struct {
		UUID         string `json:"uuid"`
		Name         string `json:"name"`
		Masters      int    `json:"masters"`
		Workers      int    `json:"workers"`
		ExternalEtcd bool   `json:"external_etc"`
		//Status string
		CreatedAt time.Time `json:"created_at"`
		CreatedBy string    `json:"created_by"`
	}

	response := []resp{}
	for _, i := range clusters {
		response = append(response, resp{UUID: i.UUID, Name: i.Name, Masters: i.Master,
			Workers: i.Worker, ExternalEtcd: i.ExternalEtcd, CreatedAt: i.CreatedAt, CreatedBy: i.CreatedBy})
	}

	json.NewEncoder(w).Encode(response)
	return
}

// ApiCluster - a local version of models.Cluster
type ApiCluster struct {
	Cluster models.Cluster
}

// CreateCluster - creates a new k8s cluster
func CreateCluster(w http.ResponseWriter, r *http.Request) {
	authOpts := context.Get(r, "authOpts").(models.AuthOpts)
	projectid := context.Get(r, "projectid")
	username := context.Get(r, "username")

	decoder := json.NewDecoder(r.Body)
	c := ApiCluster{}

	err := decoder.Decode(&c.Cluster)
	if err != nil {
		log.Error("Error decoding json for new cluster create: ", err)
		http.Error(w, "Error decoding the json data in request", 500)
		return
	}

	fmt.Println("got cluster created -> ", c.Cluster)

	c.Cluster.Master = 3
	c.Cluster.UUID = uuid.New()
	c.Cluster.CreatedAt = time.Now()

	if c.Cluster.Name == "" {
		c.Cluster.Name = "NewCluster"
	}
	c.Cluster.ProjectId = projectid.(string)
	c.Cluster.CreatedBy = username.(string)
	c.Cluster.Status = "Building"

	err = db.CreateNewCluster(&c.Cluster)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error creating cluster in the db", 500)
		return
	}

	client, err := GetComputeServcie(authOpts)
	if err != nil {
		http.Error(w, "Error creating client for openstack service", 500)
		return
	}

	// Set OS client for the cluster
	c.Cluster.OSClient = client

	// create the api lb first
	go c.CreateLB(authOpts)

	for i := 1; i <= c.Cluster.Master; i++ {
		masterNode, err := CreateVM(c.Cluster.Name, "master", i, authOpts)
		if err != nil {
			http.Error(w, "Error creating VMs for cluster", 500)
		}
		c.Cluster.MasterNodes = append(c.Cluster.MasterNodes, masterNode)
	}

	for i := 1; i <= c.Cluster.Worker; i++ {
		workerNode, err := CreateVM(c.Cluster.Name, "worker", i, authOpts)
		if err != nil {
			http.Error(w, "Error creating VMs for cluster", 500)
		}
		c.Cluster.WorkerNodes = append(c.Cluster.WorkerNodes, workerNode)
	}

	go c.goRunClusterSetup(authOpts)
	return
}

func (c *ApiCluster) goRunClusterSetup(authOpts models.AuthOpts) {
	c.TrackVMBuild(authOpts)
	c.AttachFirstMaster(authOpts)
	c.RunDeploy(authOpts)
	c.AttachMastersToLB(authOpts)
}

// UpdateCluster - update a cluster.
func UpdateCluster(w http.ResponseWriter, r *http.Request) {
}

// DeleteCluster - delete a given cluster.
func DeleteCluster(w http.ResponseWriter, r *http.Request) {
	projectid := context.Get(r, "projectid").(string)
	authOpts := context.Get(r, "authOpts").(models.AuthOpts)
	vars := mux.Vars(r)
	cluster := vars["cluster"]

	dbCluster, err := db.GetCluster(projectid, cluster)
	if err != nil {
		fmt.Println(err)
		return
	}

	// first delete all nodes
	for _, worker := range dbCluster.WorkerNodes {
		DeleteVM(worker.UUID, authOpts)
	}

	for _, master := range dbCluster.MasterNodes {
		DeleteVM(master.UUID, authOpts)
	}

	for _, etcd := range dbCluster.EtcdNodes {
		DeleteVM(etcd.UUID, authOpts)
	}

	// delete cloud lb
	if dbCluster.LBNode != nil {
		deleteLoadbalancer(dbCluster.LBNode, authOpts)
	}
	//deleteLoadbalancer(dbCluster.LBNode, authOpts)

	// update dbCluster as deleted in db
	dbCluster.Deleted = 1
	err = db.UpdateCluster(dbCluster)
	if err != nil {
		// this is bad
		fmt.Println("*** Could not find active cluster in db. ***")
		return
	}

}

// GetClusterNodes - get k8s cluster nodes.
func GetClusterNodes(w http.ResponseWriter, r *http.Request) {
}

// DeleteClusterNode - get k8s cluster node.
func DeleteClusterNode(w http.ResponseWriter, r *http.Request) {
}

func (c *ApiCluster) CreateLB(authOpts models.AuthOpts) error {
	lb := loadbalancers.LoadBalancer{}
	lb.Name = c.Cluster.Name + "-k8s-lb"
	lb.Protocol = "HTTPS"
	lb.Port = 6443

	raxlb, err := createLoadbalancer(lb, authOpts)
	if err != nil {
		return err
	}

	startTime := time.Now()

	for {
		lb := getLoadbalancer(raxlb, authOpts)
		if lb.Status == "ACTIVE" {
			c.Cluster.LBNode = lb
			break
		}

		time.Sleep(20 * time.Second)
		now := time.Now()
		if now.Sub(startTime).Minutes() > float64(10) {
			log.Println("LB not active after 10m of build... quitting nodes attach.")
			// lb did not come online, break out of the loop
			break
		}
	}

	return nil
}

func (c *ApiCluster) AttachFirstMaster(authOpts models.AuthOpts) {
	if c.Cluster.LBNode.Status != "ACTIVE" {
		fmt.Println("Error: LB not ready. LB status is not ACTIVE.")
		return
	}

	n := []string{}

	for _, m := range c.Cluster.MasterNodes {
		servername := fmt.Sprintf("k8s-%s-master-1", c.Cluster.Name)
		if m.Name == servername {
			n = append(n, m.IP)
			attachNodesToLoadbalancer(c.Cluster.LBNode, n, authOpts)
			break
		}
	}
}

func (c *ApiCluster) AttachMastersToLB(authOpts models.AuthOpts) {
	n := []string{}

	for _, m := range c.Cluster.MasterNodes {
		servername := fmt.Sprintf("k8s-%s-master-1", c.Cluster.Name)
		if m.Name != servername {
			n = append(n, m.IP)
		}
	}

	attachNodesToLoadbalancer(c.Cluster.LBNode, n, authOpts)
}

func isActive(c *ApiCluster, server string) bool {
	s, err := servers.Get(c.Cluster.OSClient, server).Extract()
	fmt.Println("Checking server status for server ", s.Name)
	if err != nil {
		fmt.Println("Cant get server status from API")
	}

	if s.Status == "ACTIVE" {
		return true
	}
	fmt.Println(s.Name, " not active yet...")
	return false
}

func (c *ApiCluster) SetNodeFacts() {
	for _, node := range c.Cluster.MasterNodes {
		s, err := servers.Get(c.Cluster.OSClient, node.UUID).Extract()
		if err != nil {
			fmt.Println("Cant get server status from API")
			continue
		}
		node.Name = s.Name
		node.IP = s.AccessIPv4
		node.InternalIP = ""
		node.Roles = []string{"master", "controlplane"}
	}

	for _, node := range c.Cluster.EtcdNodes {
		s, err := servers.Get(c.Cluster.OSClient, node.UUID).Extract()
		if err != nil {
			fmt.Println("Cant get server status from API")
			continue
		}
		node.Name = s.Name
		node.IP = s.AccessIPv4
		node.InternalIP = ""
		node.Roles = []string{"etcd", "controlplane"}
	}

	for _, node := range c.Cluster.WorkerNodes {
		s, err := servers.Get(c.Cluster.OSClient, node.UUID).Extract()
		if err != nil {
			fmt.Println("Cant get server status from API")
			continue
		}
		node.Name = s.Name
		node.IP = s.AccessIPv4
		node.InternalIP = ""
		node.Roles = []string{"worker"}
	}

	// save the cluster nodes in db
	cluster, err := db.GetCluster(c.Cluster.ProjectId, c.Cluster.UUID)
	if err != nil {
		// this is bad
		fmt.Println("*** Could not find active cluster in db. ***")
		return
	}

	cluster.MasterNodes = c.Cluster.MasterNodes
	cluster.WorkerNodes = c.Cluster.WorkerNodes
	cluster.EtcdNodes = c.Cluster.EtcdNodes
	cluster.LBNode = c.Cluster.LBNode

	err = db.UpdateCluster(cluster)
	if err != nil {
		// this is bad
		fmt.Println("*** Could not find active cluster in db. ***")
		return
	}
}

func (c *ApiCluster) TrackVMBuild(authOpts models.AuthOpts) {
	fmt.Println("Tracking vm builds ...")
	nodes := append(c.Cluster.MasterNodes, c.Cluster.WorkerNodes...)
	nodes = append(nodes, c.Cluster.EtcdNodes...)
	msg := make(chan string)

	for _, server := range nodes {
		go func(server models.Node) {
			for {
				if isActive(c, server.UUID) {
					time.Sleep(30 * time.Second)
					break
				}
				time.Sleep(30 * time.Second)
			}
			msg <- server.UUID
		}(*server)
	}

	for _, _ = range nodes {
		fmt.Println("Active servers: ", <-msg)
	}

	// At this point they are all active
	c.SetNodeFacts()
}

// RunDeploy - starts a k8s deploy and return the config if succesful
func (c *ApiCluster) RunDeploy(authOpts models.AuthOpts) (string, error) {
	// wait 60 sec to let servers settle down
	time.Sleep(60 * time.Second)

	// first deploy the fist master node
	for _, m := range c.Cluster.MasterNodes {
		fmt.Println("Master -> ", m)
		servername := fmt.Sprintf("k8s-%s-master-1", c.Cluster.Name)
		if m.Name == servername {
			sshHost := m.IP + ":22"

			cmd := fmt.Sprintf("sudo /usr/bin/kubeadm init --control-plane-endpoint '%s:6443' --upload-certs", c.Cluster.LBNode.VirtualIps[0].Address)
			fmt.Println("Running cluster init on master 1")

			out, err := runCommand(cmd, sshHost, m.Password)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			reToken := regexp.MustCompile(`--token \w+.\w+`)
			reHash := regexp.MustCompile(`--discovery-token-ca-cert-hash \w+:\w+`)
			reCert := regexp.MustCompile(`--certificate-key \w+`)

			kubeadmToken := reToken.Find([]byte(out))
			kubeadmHash := reHash.Find([]byte(out))
			kubeadmCert := reCert.Find([]byte(out))

			//fmt.Println(string(out))

			// restart kubelet after 2 mins
			go c.restartKubelet()

			for _, m := range c.Cluster.MasterNodes {
				servername := fmt.Sprintf("k8s-%s-master-1", c.Cluster.Name)
				if m.Name != servername {
					sshHost := m.IP + ":22"

					// join control plane node
					cmd := fmt.Sprintf("kubeadm join %s:6443 %s %s --control-plane %s", c.Cluster.LBNode.VirtualIps[0].Address, kubeadmToken, kubeadmHash, kubeadmCert)
					fmt.Println("join command => ", cmd)
					fmt.Println("Running join command on ", " ip ", m.IP, " password ", m.Password)

					out, err := runCommand(cmd, sshHost, m.Password)
					if err != nil {
						fmt.Println(err)
						return "", err
					}
					fmt.Println(out)
				}
			}

			// install calico
			calicoCmd := "curl https://docs.projectcalico.org/manifests/calico.yaml -o calico.yaml && /usr/bin/kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f calico.yaml"
			out, err = runCommand(calicoCmd, sshHost, m.Password)
			if err != nil {
				fmt.Println(err)
				return "", err
			}

			for _, m := range c.Cluster.WorkerNodes {
				// join control plane node
				cmd := fmt.Sprintf("kubeadm join %s:6443 %s %s", c.Cluster.LBNode.VirtualIps[0].Address, kubeadmToken, kubeadmHash)
				fmt.Println("join command => ", cmd)
				sshHost := m.IP + ":22"
				out, err := runCommand(cmd, sshHost, m.Password)
				if err != nil {
					fmt.Println(err)
					return "", err
				}
				fmt.Println(out)
			}

		}
	}

	for _, node := range c.Cluster.EtcdNodes {
		fmt.Println("Etcd -> ", node)
	}

	for _, node := range c.Cluster.WorkerNodes {
		fmt.Println("Worker -> ", node)
	}

	return "", nil
}

// this is a hack, when cluster init is done for some reason, on 2nd or 3rd node
// kubelet does not start etcd container, as a result cluster join waits for it,
// a simple kubelet restart does the trick.
func (c *ApiCluster) restartKubelet() {
	for _, m := range c.Cluster.MasterNodes {
		time.Sleep(time.Minute * 2)
		servername := fmt.Sprintf("k8s-%s-master-1", c.Cluster.Name)
		if m.Name != servername {
			sshHost := m.IP + ":22"
			// stop kubelet
			kubeletcmd := "service kubelet restart"
			_, err := runCommand(kubeletcmd, sshHost, m.Password)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func runCommand(cmd, host, pass string) (string, error) {
	client, session, err := connectToHost("root", pass, host)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	client.Close()
	return string(out), nil
}
