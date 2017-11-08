package lifecycle

import (
	"fmt"
	"os/user"
	"time"

	"github.com/mlab-lattice/kubernetes-integration/pkg/constants"
	"github.com/mlab-lattice/kubernetes-integration/pkg/util/minikube"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

type LocalProvisioner struct {
	latticeImageDockerRepository string
	mec                          *minikube.ExecContext
}

func NewLocalProvisioner(latticeImageDockerRepository, logPath string) (*LocalProvisioner, error) {
	mec, err := minikube.NewMinikubeExecContext(logPath)
	if err != nil {
		return nil, err
	}

	lp := &LocalProvisioner{
		latticeImageDockerRepository: latticeImageDockerRepository,
		mec: mec,
	}
	return lp, nil
}

func (lp *LocalProvisioner) Provision(name, url string) error {
	result, logFilename, err := lp.mec.Start(name)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube start (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*lp.mec.LogPath, logFilename))

	err = result.Wait()
	if err != nil {
		return err
	}

	address, err := lp.Address(name)
	if err != nil {
		return err
	}

	err = lp.bootstrap(address)
	if err != nil {
		return err
	}

	fmt.Println("Waiting for System Environment Manager to be ready...")
	return pollForSystemEnvironmentReadiness(address)
}

func (lp *LocalProvisioner) Address(name string) (string, error) {
	return lp.mec.IP(name)
}

func (lp *LocalProvisioner) bootstrap(address string) error {
	fmt.Println("Bootstrapping")
	usr, err := user.Current()
	if err != nil {
		return err
	}
	// TODO: support passing in the context when supported
	// https://github.com/kubernetes/minikube/issues/2100
	//configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubeContext}
	configOverrides := &clientcmd.ConfigOverrides{}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: filepath.Join(usr.HomeDir, ".kube/config")},
		configOverrides,
	).ClientConfig()

	if err != nil {
		return err
	}

	kubeClientset := clientset.NewForConfigOrDie(config)

	fmt.Println("Creating bootstrap SA")
	bootstrapSA := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubernetes-bootstrapper",
			Namespace: constants.NamespaceDefault,
		},
	}

	_, err = kubeClientset.
		CoreV1().
		ServiceAccounts(constants.NamespaceDefault).
		Create(bootstrapSA)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	bootstrapClusterAdminRoleBind := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubernetes-bootstrapper-cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      bootstrapSA.Name,
				Namespace: bootstrapSA.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}
	_, err = kubeClientset.
		RbacV1().
		ClusterRoleBindings().
		Create(&bootstrapClusterAdminRoleBind)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	jobName := "lattice-bootstrap-kubernetes"
	var backoffLimit int32 = 2
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: jobName,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "bootstrap-kubernetes",
							Image:   lp.latticeImageDockerRepository + "/" + constants.DockerImageBootstrapKubernetes,
							Command: []string{"/app/cmd/bootstrap-kubernetes/go_image.binary"},
							Args:    []string{"-provider", "local", "-user-system-url", "github.com/foo/bar", "-system-ip", address},
						},
					},
					RestartPolicy:      corev1.RestartPolicyNever,
					DNSPolicy:          corev1.DNSDefault,
					ServiceAccountName: bootstrapSA.Name,
				},
			},
		},
	}

	fmt.Println("Creating bootstrap job")
	_, err = kubeClientset.
		BatchV1().
		Jobs(constants.NamespaceDefault).
		Create(&job)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	fmt.Println("Polling bootstrap job status")
	err = wait.Poll(1*time.Second, 300*time.Second, func() (bool, error) {
		j, err := kubeClientset.BatchV1().Jobs(constants.NamespaceDefault).Get(job.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if j.Status.Succeeded == 1 {
			return true, nil
		}

		if j.Status.Failed >= backoffLimit {
			return false, fmt.Errorf("surpassed backoffLimit")
		}

		return false, nil
	})
	if err != nil {
		return err
	}

	fmt.Println("Deleting bootstrap SA")
	return kubeClientset.CoreV1().ServiceAccounts(constants.NamespaceDefault).Delete(bootstrapSA.Name, nil)
}

func (lp *LocalProvisioner) Deprovision(name string) error {
	result, logFilename, err := lp.mec.Delete(name)
	if err != nil {
		return err
	}

	fmt.Printf("Running minikube delete (pid: %v, log file: %v)\n", result.Pid, filepath.Join(*lp.mec.LogPath, logFilename))

	return result.Wait()
}
