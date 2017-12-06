package app

import (
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

func pollKubeResourceCreation(resourceCreationFunc func() (interface{}, error)) {
	err := wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
		_, err := resourceCreationFunc()

		if err != nil && !apierrors.IsAlreadyExists(err) {
			fmt.Printf("encountered error from API: %v\n", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		panic(err)
	}
}

func getContainerImageFQN(repository string) string {
	return fmt.Sprintf("%v/%v%v", latticeContainerRegistry, latticeContainerRepoPrefix, repository)
}
