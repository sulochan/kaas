package mongodb

import (
	"errors"
	"fmt"

	"github.com/sulochan/kaas/models"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const dbname = "kaas"

var mongoSession *mgo.Session

var (
	NotFound = errors.New("Not Found")
)

func init() {
	//conf := config.GetConfig()
	Msession, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	if err := Msession.Ping(); err != nil {
		panic(err)
	}
	Msession.SetMode(mgo.Monotonic, true)
	mongoSession = Msession
}

func CreateNewCluster(cluster *models.Cluster) error {
	fmt.Println("Create new cluster got called.")
	session := mongoSession.Copy()
	defer session.Close()
	coll := session.DB(dbname).C("clusters")
	err := coll.Insert(cluster)
	return err
}

func GetAllClusters(projectid string) ([]models.Cluster, error) {
	fmt.Println("Get all new cluster got called.")
	session := mongoSession.Copy()
	defer session.Close()
	clusters := []models.Cluster{}
	coll := session.DB(dbname).C("clusters")
	err := coll.Find(bson.M{"projectid": projectid, "deleted": 0}).All(&clusters)
	return clusters, err
}

func GetCluster(projectid string, uuid string) (*models.Cluster, error) {
	session := mongoSession.Copy()
	defer session.Close()
	cluster := models.Cluster{}
	coll := session.DB(dbname).C("clusters")
	err := coll.Find(bson.M{"projectid": projectid, "uuid": uuid, "deleted": 0}).One(&cluster)
	return &cluster, err
}

func UpdateCluster(cluster *models.Cluster) error {
	session := mongoSession.Copy()
	defer session.Close()
	coll := session.DB(dbname).C("clusters")
	query := bson.M{"projectid": cluster.ProjectId, "uuid": cluster.UUID, "deleted": 0}
	change := bson.M{"$set": cluster}

	err := coll.Update(query, change)
	return err
}

// RegisterNode registers a new node in the database
func RegisterNode(node *models.Node) error {
	session := mongoSession.Copy()
	defer session.Close()
	coll := session.DB(dbname).C("nodes")
	err := coll.Insert(node)
	return err
}

// GetNextCommand returns the next command to be executed on the node
func GetNextCommand(projectid string, uuid string) (*models.Job, error) {
	session := mongoSession.Copy()
	defer session.Close()
	job := models.Job{}
	coll := session.DB(dbname).C("jobs")
	err := coll.Find(bson.M{"projectid": projectid, "uuid": uuid, "deleted": 0}).One(&job)
	return &job, err
}

// UpdateJob updates the job in the database
func UpdateJob(job *models.Job) error {
	session := mongoSession.Copy()
	defer session.Close()
	coll := session.DB(dbname).C("jobs")
	query := bson.M{"uuid": job.UUID}
	change := bson.M{"$set": job}
	err := coll.Update(query, change)
	return err
}
