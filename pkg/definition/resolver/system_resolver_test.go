package resolver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/mlab-lattice/system/pkg/util/git"

	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const (
	testRepoDir     = "/tmp/lattice-core/test/resolver/my-repo"
	testRepoURI1    = "file:///tmp/lattice-core/test/resolver/my-repo/.git#v1"
	testRepoURI2    = "file:///tmp/lattice-core/test/resolver/my-repo/.git#v2"
	testWorkDir     = "/tmp/lattice-core/test/resolver/work"
	systemFileName  = "system.json"
	serviceFileName = "service.json"
)

func TestMain(m *testing.M) {
	fmt.Println("Running resolvers tests...")
	setupTest()
	retCode := m.Run()
	teardownTest()
	os.Exit(retCode)
}

func TestValidateSystemResolver(t *testing.T) {

	testV1(t)
	testV2(t)
	testListVersions(t)
}

func testV1(t *testing.T) {
	fmt.Println("--------------- Testing ResolveDefinition V1")

	res, err := NewSystemResolver(testWorkDir)
	if err != nil {
		t.Fatalf("Got error calling NewSystemResolver: %v", err)
	}

	defNode, err := res.ResolveDefinition(path.Join(testRepoURI1, "system.json"), nil)
	if err != nil {
		t.Fatalf("Error is not nil: %v", err)
	}

	if defNode.Name() != "my-system-v1" {
		t.Error("Wrong system name")
	}

	if len(defNode.Subsystems()) != 2 {
		t.Error("Wrong # of subsystems")
	}

	if len(defNode.Services()) != 2 {
		t.Error("Wrong # of services")
	}

	if defNode.Subsystems()["/my-system-v1/my-service"].Name() != "my-service" {
		t.Error("Invalid Subsystem map")
	}

}

func testV2(t *testing.T) {

	fmt.Println("--------------- Testing ResolveDefinition V2")

	res, err := NewSystemResolver(testWorkDir)
	if err != nil {
		t.Fatalf("Got error calling NewSystemResolver: %v", err)
	}

	defNode, err := res.ResolveDefinition(path.Join(testRepoURI2, "system.json"), nil)
	if err != nil {
		t.Error("Error is not nil: ", err)
	}

	if defNode.Name() != "my-system-v2" {
		t.Error("Wrong system name")
	}
}

func testListVersions(t *testing.T) {
	fmt.Println("--------------- Testing ListDefinitionVersions")

	res, err := NewSystemResolver(testWorkDir)
	if err != nil {
		t.Fatalf("Got error calling NewSystemResolver: %v", err)
	}

	versions, err := res.ListDefinitionVersions(testRepoURI2, &git.Options{})
	if err != nil {
		t.Fatal("Error is not nil: ", err)
	}

	if len(versions) != 2 {
		t.Error("Wrong # of versions")
	}

	if versions[0] != "v1" || versions[1] != "v2" {
		t.Error("Wrong version")
	}

}

func setupTest() {

	fmt.Println("Setting up resolver test")
	// ensure work directory
	os.Mkdir(testRepoDir, 0700)

	gogit.PlainInit(testRepoDir, false)

	commitTestFiles(SYSTEM_JSON, serviceJSON, "v1")
	commitTestFiles(systemJSON2, serviceJSON, "v2")

}

func commitTestFiles(systemJson string, serviceJson string, tag string) {

	systemFileContents := []byte(systemJson)
	ioutil.WriteFile(path.Join(testRepoDir, systemFileName), systemFileContents, 0644)

	serviceFileContents := []byte(serviceJson)
	ioutil.WriteFile(path.Join(testRepoDir, serviceFileName), serviceFileContents, 0644)

	repo, _ := gogit.PlainOpen(testRepoDir)

	workTree, _ := repo.Worktree()

	workTree.Add(systemFileName)

	workTree.Add(serviceFileName)

	// commit
	hash, _ := workTree.Commit("test", &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@mlab-lattice.com",
			When:  time.Now(),
		},
	})

	// create the tag
	n := plumbing.ReferenceName("refs/tags/" + tag)
	t := plumbing.NewHashReference(n, hash)
	repo.Storer.SetReference(t)

}

func teardownTest() {
	fmt.Println("Tearing down resolver test")
	os.RemoveAll(testRepoDir)
	os.RemoveAll(testWorkDir)

}

const SYSTEM_JSON = `
{
  "name": "my-system-v1",
  "type": "system",
  "subsystems": [
    {"$include": "service.json"},
    {
      "name": "my-service",
      "type": "service",
      "components": [
        {
          "name": "service",
          "build": {
            "docker_image": {
              "registry": "registry.company.com",
              "repository": "foobar",
              "tag": "v1.0.0"
            }
          },
          "exec": {
            "command": [
              "./start",
              "--my-app"
            ],
            "environment": {
              "biz": "baz",
              "foo": "bar"
            }
          }
        }
      ],
      "resources": {
        "min_instances": 1,
        "max_instances": 1,
        "instance_type": "mock.instance.type"
      }
    }
  ]
}
`

const serviceJSON = `
{
  "name": "my-service-2",
  "type": "service",
  "components": [
    {
      "name": "service",
      "build": {
        "docker_image": {
          "registry": "registry.company.com",
          "repository": "foobar",
          "tag": "v1.0.0"
        }
      },
      "exec": {
        "command": [
          "./start",
          "--my-app"
        ],
        "environment": {
          "biz": "baz",
          "foo": "bar"
        }
      }
    }
  ],
  "resources": {
    "min_instances": 1,
    "max_instances": 1,
    "instance_type": "mock.instance.type"
  }
}
`

const systemJSON2 = `
{
  "name": "my-system-v2",
  "type": "system",
  "subsystems": [
    {"$include": "service.json"},
    {
	  "name": "my-service",
      "type": "service",
      "components": [
        {
          "name": "service",
          "build": {
            "docker_image": {
              "registry": "registry.company.com",
              "repository": "foobar",
              "tag": "v1.0.0"
            }
          },
          "exec": {
            "command": [
              "./start",
              "--my-app"
            ],
            "environment": {
              "biz": "baz",
              "foo": "bar"
            }
          }
        }
      ],
      "resources": {
        "min_instances": 1,
        "max_instances": 1,
        "instance_type": "mock.instance.type"
      }
    }
  ]
}
`
