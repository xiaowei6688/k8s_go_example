package main

import (
	"context"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"log"
	"path/filepath"
	"time"
)

// 实现一个简单的功能，创建一个deployment
func createDeployment(dpClient v1.DeploymentInterface) error {
	replicas := int32(3)
	newDp := appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-deployment",
		},
		Spec: appv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					// 这里的label要和上面的selector的label一致
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							// 这里的name要和下面的container的name一致
							Name:  "nginx",
							Image: "nginx:1.16",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := dpClient.Create(context.TODO(), &newDp, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// 修改deployment
func updateDeployment(dpClient v1.DeploymentInterface) error {
	dp, err := dpClient.Get(context.TODO(), "nginx-deployment", metav1.GetOptions{})
	if err != nil {
		return err
	}
	dp.Spec.Template.Spec.Containers[0].Image = "nginx:1.17"
	updateDp, err := dpClient.Update(context.TODO(), dp, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// 通过retry.RetryOnConflict来解决更新冲突
		_, err := dpClient.Update(context.TODO(), updateDp, metav1.UpdateOptions{})
		return err
	})
}

// 删除deployment
func deleteDeployment(dpClient v1.DeploymentInterface) error {
	deletePolicy := metav1.DeletePropagationForeground
	err := dpClient.Delete(context.TODO(), "nginx-deployment", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	homePath := homedir.HomeDir()
	if homePath == "" {
		panic("homePath is empty")
	}
	kubeConfig := filepath.Join(homePath, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err)
	}

	// use the config to create a client
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	dpClient := clientSet.AppsV1().Deployments(corev1.NamespaceDefault)

	log.Println("start handle deployments...")

	// 创建一个deployment
	err = createDeployment(dpClient)
	if err != nil {
		panic(err)
	}
	log.Println("create deployment success")

	<-time.Tick(time.Minute * 1)

	// 修改一个deployment
	err = updateDeployment(dpClient)
	if err != nil {
		panic(err)
	}
	log.Println("update deployment success")

	<-time.Tick(time.Minute * 1)
	
	// 删除一个deployment
	err = deleteDeployment(dpClient)
	if err != nil {
		panic(err)
	}
	log.Println("delete deployment success")
}
