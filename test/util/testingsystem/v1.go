package testingsystem

import (
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system"
	"github.com/mlab-lattice/system/test/util/versionservice"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/gomega"
)

const (
	V1ServiceAVersion          = "1.0.0"
	V1ServiceAPath             = tree.NodePath("/test/a")
	V1ServiceAPublicPort int32 = 8080
)

type V1 struct {
	systemID v1.SystemID
	v1client v1client.Interface
}

func NewV1(client v1client.Interface, systemID v1.SystemID) *V1 {
	return &V1{
		systemID: systemID,
		v1client: client,
	}
}

func (v *V1) ValidateStable() {
	sys := system.Get(v.v1client.Systems(), v.systemID)

	Expect(sys.State).To(Equal(v1.SystemStateStable))

	Expect(len(sys.Services)).To(Equal(1))
	service, ok := sys.Services[V1ServiceAPath]
	Expect(ok).To(BeTrue())

	Expect(service.State).To(Equal(v1.ServiceStateStable))
	Expect(service.StaleInstances).To(Equal(int32(0)))
	Expect(service.UpdatedInstances).To(Equal(int32(1)))
	Expect(len(service.PublicPorts)).To(Equal(1))
	port, ok := service.PublicPorts[V1ServiceAPublicPort]
	Expect(ok).To(BeTrue())

	err := v.poll(port.Address, time.Second, 30*time.Second)
	Expect(err).To(Not(HaveOccurred()))
}

func (v *V1) test(serviceAURL string) error {
	client := versionservice.NewClient(serviceAURL)
	return client.CheckStatusAndVersion(V1ServiceAVersion)
}

func (v *V1) poll(serviceAURL string, interval, timeout time.Duration) error {
	err := wait.Poll(interval, timeout, func() (bool, error) {
		return false, v.test(serviceAURL)
	})
	if err == nil || err == wait.ErrWaitTimeout {
		return nil
	}

	return err
}
